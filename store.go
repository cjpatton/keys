// Copyright (c) 2017, Christopher Patton.
// All rights reserved.
package store

// TODO(me) Rename GetRow.

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"unsafe"

	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/pbkdf2"
)

/*
// The next line gets things going on Mac:
#cgo CPPFLAGS: -I/usr/local/opt/openssl/include
#cgo LDFLAGS: -lstruct -lcrypto
#include <struct/const.h>
#include <struct/dict.h>
#include "string.h"

char **new_str_list(int len) {
	return calloc(sizeof(char *), len);
}

int *new_int_list(int len) {
	return calloc(sizeof(int), len);
}

void set_str_list(char **list, int idx, char *val) {
	list[idx] = val;
}

void set_int_list(int *list, int idx, int val) {
	list[idx] = val;
}

int get_int_list(int *list, int idx) {
	return list[idx];
}

void free_str_list(char **list, int len) {
	int i;
	for (i = 0; i < len; i++) {
		if (list[i] != NULL) {
			free(list[i]);
		}
	}
	free(list);
}

void free_int_list(int *list) {
	free(list);
}

char *get_row_ptr(char *table, int row, int row_bytes) {
	return &table[row * row_bytes];
}
*/
import "C"

// Number of bytes to use for the salt. The salt is a random string used to
// construct the table. It is prepended to the input of each HMAC call.
const SaltBytes = 8

// Number of row bytes allocated for the tag.
const TagBytes = 3

// The maximum length of the row. In general, the length of the row depends on
// the length of the longest output in the map. HASH_BYTES is defined in
// c/const.h.
const MaxRowBytes = C.HASH_BYTES

// The maximum length of the outputs. 1 byte of each row is allocated for
// padding the output string.
const MaxOutputBytes = MaxRowBytes - TagBytes - 1

// Length of the HMAC key. HMAC_KEY_BYTES is defined in c/const.h.
const KeyBytes = C.HMAC_KEY_BYTES

type Error string

func (err Error) Error() string {
	return string(err)
}

// Returned by Get() and priv.GetValue() if the input was not found in the
// dictionary.
const ItemNotFound = Error("item not found")

// Returned by GetRow() in case idx not in the table index.
const ErrorIdx = Error("index out of range")

// CError propagates an error from the internal C code.
func CError(fn string, errNo C.int) Error {
	return Error(fmt.Sprintf("%s returns error %d", fn, errNo))
}

// The public representation of the map.
type PubStore struct {
	dict *C.dict_t
}

// The private state required for evaluation queries.
type PrivStore struct {
	tinyCtx    *C.tiny_ctx
	params     C.dict_params_t
	cZeroShare *C.char
}

// GenerateKey generates a fresh, random key and returns it.
func GenerateKey() []byte {
	K := make([]byte, KeyBytes)
	_, err := rand.Read(K)
	if err != nil {
		return nil
	}
	return K
}

// DeriveKeyFromPassword derives a key from a password and (optional) salt and
// returns it.
//
// NOTE The salt is not the same as StoreParams.Salt. StoreParams.Salt is
// generated by New(), which in turn depends on the key.
func DeriveKeyFromPassword(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 4096, KeyBytes, sha256.New)
}

// New generates a new structure (pub, priv) for the map M and key K.
//
// NOTE You must call pub.Free() and priv.Free() before these variables go out
// of scope. These structures contain C types that were allocated on the heap
// and must be freed before losing a reference to them.
func New(K []byte, M map[string]string) (*PubStore, *PrivStore, error) {

	pub := new(PubStore)

	// Copy input/output pairs into C land.
	itemCt := C.int(len(M))
	inputs := C.new_str_list(itemCt)
	inputBytes := C.new_int_list(itemCt)
	outputs := C.new_str_list(itemCt)
	outputBytes := C.new_int_list(itemCt)
	defer C.free_str_list(inputs, itemCt)
	defer C.free_str_list(outputs, itemCt)
	defer C.free_int_list(inputBytes)
	defer C.free_int_list(outputBytes)

	maxOutputueBytes := 0
	i := C.int(0)
	for input, output := range M {
		if len(output) > maxOutputueBytes {
			maxOutputueBytes = len(output)
		}
		// NOTE C.CString() copies all the bytes of its input, even if it
		// encounters a null byte.
		C.set_str_list(inputs, i, C.CString(input))
		C.set_int_list(inputBytes, i, C.int(len(input)))
		C.set_str_list(outputs, i, C.CString(output))
		C.set_int_list(outputBytes, i, C.int(len(output)))
		i++
	}

	// Allocate a new dictionary object.
	tableLen := C.dict_compute_table_length(C.int(len(M)))
	pub.dict = C.dict_new(
		tableLen,
		C.int(maxOutputueBytes),
		C.int(TagBytes),
		C.int(SaltBytes))
	if pub.dict == nil {
		return nil, nil, Error(fmt.Sprintf("maxOutputBytes > %d", MaxOutputBytes))
	}

	params := cParamsToStoreParams(&pub.dict.params)

	// Create priv.
	//
	// NOTE dict.salt is not set, and so priv.params.salt is not set. It's
	// necessary to set it after calling C.dict_create().
	priv, err := NewPrivStore(K, params)
	if err != nil {
		return nil, nil, err
	}

	// Create the dictionary.
	errNo := C.dict_create(
		pub.dict, priv.tinyCtx, inputs, inputBytes, outputs, outputBytes, itemCt)
	if errNo != C.OK {
		priv.Free()
		return nil, nil, CError("dict_create", errNo)
	}

	// Copy salt to priv.params.
	C.memcpy(unsafe.Pointer(priv.params.salt),
		unsafe.Pointer(pub.dict.params.salt),
		C.size_t(priv.params.salt_bytes))

	return pub, priv, nil
}

// NewPubStoreFromTable creates a new *PubStore from a *StoreTable protobuf.
//
// NOTE You must destrohy the output with pub.Free().
func NewPubStoreFromTable(table *StoreTable) *PubStore {
	pub := new(PubStore)
	pub.dict = (*C.dict_t)(C.malloc(C.sizeof_dict_t))

	// Allocate memory for salt + 1 tweak byte and set the parameters.
	pub.dict.params.salt = (*C.char)(C.malloc(C.size_t(len(table.GetParams().Salt) + 1)))
	setCParamsFromStoreParams(&pub.dict.params, table.GetParams())

	// Allocate memory for table + 1 zero row and copy the table.
	tableLen := C.int(table.GetParams().GetTableLen())
	rowBytes := C.int(table.GetParams().GetRowBytes())
	realTableLen := C.int(len(table.Table)) / rowBytes
	cBuf := C.CString(string(table.Table))
	defer C.free(unsafe.Pointer(cBuf))
	pub.dict.table = (*C.char)(C.malloc(C.size_t(tableLen * rowBytes)))
	C.memset(unsafe.Pointer(pub.dict.table), 0, C.size_t(tableLen*rowBytes))
	for i := 0; i < int(realTableLen); i++ {
		src := C.get_row_ptr(cBuf, C.int(i), rowBytes)
		dst := C.get_row_ptr(pub.dict.table, C.int(table.Idx[i]), rowBytes)
		C.memcpy(unsafe.Pointer(dst), unsafe.Pointer(src), C.size_t(rowBytes))
	}

	return pub
}

// Get queries input on the structure (pub, priv). The result is M[input] =
// output, where M is the map represented by (pub, priv).
func Get(pub *PubStore, priv *PrivStore, input string) (string, error) {
	cInput := C.CString(input)

	// NOTE(me) Better way to do the following?
	cOutput := C.CString(string(make([]byte, pub.dict.params.max_value_bytes)))
	cOutputBytes := C.int(0)
	defer C.free(unsafe.Pointer(cInput))
	defer C.free(unsafe.Pointer(cOutput))
	errNo := C.dict_get(
		pub.dict, priv.tinyCtx, cInput, C.int(len(input)), cOutput, &cOutputBytes)
	if errNo == C.ERR_DICT_BAD_KEY {
		return "", ItemNotFound
	} else if errNo != C.OK {
		return "", CError("cdict_get", errNo)
	}
	return C.GoStringN(cOutput, cOutputBytes), nil
}

// GetShare returns the bitwise-XOR of the x-th and y-th rows of the table.
func (pub *PubStore) GetShare(x, y int) ([]byte, error) {
	if x < 0 || x >= int(pub.dict.params.table_length) ||
		y < 0 || y >= int(pub.dict.params.table_length) {
		return nil, ErrorIdx
	}
	xRow := getRow(pub.dict.table, C.int(x), pub.dict.params.row_bytes)
	yRow := getRow(pub.dict.table, C.int(y), pub.dict.params.row_bytes)
	for i := 0; i < len(xRow); i++ {
		xRow[i] ^= yRow[i]
	}
	return xRow, nil
}

// ToString returns a string representation of the table.
func (pub *PubStore) ToString() string {
	return pub.GetTable().String()
}

// GetTable returns a *StoreTable protobuf representation of the dictionary.
func (pub *PubStore) GetTable() *StoreTable {
	cdict := C.dict_compress(pub.dict)
	defer C.cdict_free(cdict)
	rowBytes := int(pub.dict.params.row_bytes)
	tableLen := int(cdict.compressed_table_length)
	tableIdx := make([]int32, tableLen)
	for i := 0; i < tableLen; i++ {
		tableIdx[i] = int32(C.get_int_list(cdict.idx, C.int(i)))
	}
	return &StoreTable{
		Params: cParamsToStoreParams(&pub.dict.params),
		Table:  C.GoBytes(unsafe.Pointer(cdict.table), C.int(tableLen*rowBytes)),
		Idx:    tableIdx,
	}
}

// Free deallocates memory associated with the underlying C implementation of
// the data structure.
func (pub *PubStore) Free() {
	C.dict_free(pub.dict)
}

// NewPrivStore creates a new *PrivStore from a key and parameters.
//
// NOTE You must destroy this with priv.Free().
// NOTE Called by New().
func NewPrivStore(K []byte, params *StoreParams) (*PrivStore, error) {
	priv := new(PrivStore)

	// Check that K is the right length.
	if len(K) != KeyBytes {
		return nil, Error(fmt.Sprintf("len(K) = %d, expected %d", len(K), KeyBytes))
	}

	// Create new tinyprf context.
	priv.tinyCtx = C.tinyprf_new(C.int(params.GetTableLen()))
	if priv.tinyCtx == nil {
		return nil, Error("tableLen < 2")
	}

	// Allocate memory for salt.
	priv.params.salt = (*C.char)(C.malloc(C.size_t(len(params.Salt) + 1)))

	// Initialize tinyprf.
	cK := C.CString(string(K))
	defer C.memset(unsafe.Pointer(cK), 0, C.size_t(KeyBytes))
	defer C.free(unsafe.Pointer(cK))
	errNo := C.tinyprf_init(priv.tinyCtx, cK)
	if errNo != C.OK {
		priv.Free()
		return nil, CError("tinyprf_init", errNo)
	}

	// Set parameters.
	setCParamsFromStoreParams(&priv.params, params)

	// A 0-byte string used by GetValue().
	priv.cZeroShare = (*C.char)(C.malloc(C.size_t(priv.params.row_bytes)))
	C.memset(unsafe.Pointer(priv.cZeroShare), 0, C.size_t(priv.params.row_bytes))

	return priv, nil
}

// GetIdx computes the two indices of the table associated with input and
// returns them.
func (priv *PrivStore) GetIdx(input string) (int, int, error) {
	cInput := C.CString(input)
	defer C.free(unsafe.Pointer(cInput))
	var x, y C.int
	errNo := C.dict_compute_rows(
		priv.params, priv.tinyCtx, cInput, C.int(len(input)), &x, &y)
	if errNo != C.OK {
		return 0, 0, CError("dict_compute_rows", errNo)
	}
	return int(x), int(y), nil
}

// GetValue computes the output associated with the input and the table rows.
func (priv *PrivStore) GetValue(input string, pubShare []byte) (string, error) {
	cInput := C.CString(input)
	// NOTE(me) Better way to do the following?
	cOutput := C.CString(string(make([]byte, priv.params.max_value_bytes)))
	defer C.free(unsafe.Pointer(cInput))
	defer C.free(unsafe.Pointer(cOutput))
	cOutputBytes := C.int(0)

	cPubShare := C.CString(string(pubShare))
	defer C.free(unsafe.Pointer(cPubShare))

	errNo := C.dict_compute_value(priv.params, priv.tinyCtx, cInput,
		C.int(len(input)), cPubShare, priv.cZeroShare, cOutput, &cOutputBytes)

	if errNo == C.ERR_DICT_BAD_KEY {
		return "", ItemNotFound
	} else if errNo != C.OK {
		return "", CError("dict_compute_value", errNo)
	}
	return C.GoStringN(cOutput, cOutputBytes), nil
}

// GetParams returns the public parameters of the data structure.
func (priv *PrivStore) GetParams() *StoreParams {
	return cParamsToStoreParams(&priv.params)
}

// Free deallocates moemory associated with the C implementation of the
// underlying data structure.
func (priv *PrivStore) Free() {
	C.free(unsafe.Pointer(priv.params.salt))
	C.free(unsafe.Pointer(priv.cZeroShare))
	C.tinyprf_free(priv.tinyCtx)
}

// Returns true if the first saltBytes of *a and *b are equal.
func cBytesToString(str *C.char, bytes C.int) string {
	return C.GoStringN(str, bytes)
}

// cParamsToStoreParams creates *StoreParams from a *C.dict_params_t, making a
// deep copy of the salt.
//
// Called by pub.GetParams() and priv.GetParams().
func cParamsToStoreParams(cParams *C.dict_params_t) *StoreParams {
	return &StoreParams{
		TableLen:       *proto.Int32(int32(cParams.table_length)),
		MaxOutputBytes: *proto.Int32(int32(cParams.max_value_bytes)),
		RowBytes:       *proto.Int32(int32(cParams.row_bytes)),
		TagBytes:       *proto.Int32(int32(cParams.tag_bytes)),
		Salt:           C.GoBytes(unsafe.Pointer(cParams.salt), cParams.salt_bytes),
	}
}

// setCParamsFromStoreparams copies parameters to a *C.dict_params_t.
//
// Must call C.free(cParams.salt)
func setCParamsFromStoreParams(cParams *C.dict_params_t, params *StoreParams) {
	cParams.table_length = C.int(params.GetTableLen())
	cParams.max_value_bytes = C.int(params.GetMaxOutputBytes())
	cParams.row_bytes = C.int(params.GetRowBytes())
	cParams.tag_bytes = C.int(params.GetTagBytes())
	cParams.salt_bytes = C.int(len(params.Salt))
	cBuf := C.CString(string(params.Salt))
	C.memcpy(unsafe.Pointer(cParams.salt),
		unsafe.Pointer(cBuf),
		C.size_t(cParams.salt_bytes))
}

// getRow returns a []byte corresponding to row in the table.
func getRow(table *C.char, idx, rowBytes C.int) []byte {
	rowPtr := C.get_row_ptr(table, idx, rowBytes)
	return C.GoBytes(unsafe.Pointer(rowPtr), rowBytes)
}

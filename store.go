// Copyright (c) 2017, Christopher Patton
// All rights reserved.

package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/cjpatton/store/pb"
	"golang.org/x/crypto/pbkdf2"
)

// Length of key for sealing the outputs. (AEAD is AES128-GCM.)
const SealKeyBytes = 16

// Length of the store key.
const KeyBytes = DictKeyBytes + SealKeyBytes

// Returned by NewStore() in case the number of elements in the input exceeds
// the number of unique counters.
const ErrorMapTooLarge = Error("input map is too large")

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
// If salt == nil, then no salt is used. Note that the salt is not the same as
// pb.Params.Salt. pb.Params.Salt is generated by NewDict(), which in turn
// depends on the key.
func DeriveKeyFromPassword(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 4096, KeyBytes, sha256.New)
}

// Stores the public representation of the map.
type PubStore struct {
	dict   *PubDict
	sealed [][]byte
	g      graph
}

// Stores the private context used to query the map.
type PrivStore struct {
	dict *PrivDict
	aead cipher.AEAD
}

// NewStore creates a new store for key K and map M.
//
// You must call pub.Free() and priv.Free() before these variables go out of
// scope. This is necessary because these structures contain memory allocated
// from the heap in C.
func NewStore(K []byte, M map[string]string) (pub *PubStore, priv *PrivStore, err error) {

	pub = new(PubStore)
	priv = new(PrivStore)

	// Set up context for AEAD.
	block, err := aes.NewCipher(K[:DictKeyBytes])
	if err != nil {
		return nil, nil, err
	}
	priv.aead, err = cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// AEAD nonce is the dictionary salt plus a counter.
	//
	// Compute the number of bytes of the nonce allocated for the counter and
	// ensure that it is long enough to uniquely encode each input/output pair
	// in the map.
	ctrBytes := priv.aead.NonceSize() - SaltBytes
	if len(M) > (1 << (8 * uint(ctrBytes))) {
		return nil, nil, ErrorMapTooLarge
	}

	inputs := make([][]byte, len(M))
	outputs := make([][]byte, len(M))
	i := 0
	for in, out := range M {
		inputs[i] = []byte(in)
		outputs[i] = []byte(out)
		i++
	}

	// Create a *cMap for inputs to counters. This is what will actually be
	// stored by pub.dict.
	cN := newCtrCMap(inputs, ctrBytes)
	defer cN.free()

	// Construct the graph.
	pub.dict, priv.dict, pub.g, err = newDictAndGraph(
		K[DictKeyBytes:], cN, 0, false)
	if err != nil {
		return nil, nil, err
	}

	// Encrypt each output and store in pub.sealed.
	nonce := cBytesToBytes(priv.dict.params.salt, priv.dict.params.salt_bytes)
	ctr := make([]byte, ctrBytes)
	pub.sealed = make([][]byte, len(M))
	for i := 0; i < len(M); i++ {
		binary.LittleEndian.PutUint32(ctr, uint32(i))
		pub.sealed[i] = priv.aead.Seal(nil, append(nonce, ctr...),
			outputs[i], inputs[i])
	}

	return pub, priv, nil
}

// GetIdx computes the index corresponding to the input.
func (priv *PrivStore) GetIdx(input string) (int, int, error) {
	return priv.dict.GetIdx(input)
}

// GetShare computes the pubShare corresponding to the index (x, y).
//
// The payload is comprised of the public share of the dictionary table and the
// sealed output corresponding to the share. This function returns ItemNotFound
// if there is no such sealed output.
func (pub *PubStore) GetShare(x, y int) ([]byte, error) {
	// Get counter share.
	ctrShare, err := pub.dict.GetShare(x, y)
	if err != nil {
		return nil, err
	}
	// Look up sealed output.
	for i := 0; i < len(pub.g[x]); i++ {
		e := pub.g[x][i]
		for j := 0; j < len(pub.g[y]); j++ {
			if pub.g[y][j] == e {
				return append(ctrShare, pub.sealed[e]...), nil
			}
		}
	}
	return nil, ItemNotFound
}

// GetOutput computes the final output from input and the public share.
//
// The nonce is the constructed from combining the table public share with the
// private share and concatenating the result to the salt. The associated data
// is the input. Returns ItemNotFound if unsealing the output fails.
func (priv *PrivStore) GetOutput(input string, pubShare []byte) (string, error) {
	ctrShareBytes := priv.dict.params.row_bytes
	ctr, err := priv.dict.GetOutput(input, pubShare[:ctrShareBytes])
	if err != nil {
		return "", err
	}

	nonce := cBytesToBytes(priv.dict.params.salt, priv.dict.params.salt_bytes)
	output, err := priv.aead.Open(
		nil, append(nonce, []byte(ctr)...), pubShare[ctrShareBytes:], []byte(input))
	if err != nil {
		return "", ItemNotFound
	}
	return string(output), nil
}

// Get looks up input in the public store and returns the result.
func (priv *PrivStore) Get(pub *PubStore, input string) (string, error) {
	x, y, err := priv.GetIdx(input)
	if err != nil {
		return "", err
	}

	pubShare, err := pub.GetShare(x, y)
	if err != nil {
		return "", err
	}

	return priv.GetOutput(input, pubShare)
}

// Free releases memory allocated to the public store's internal representation.
func (pub *PubStore) Free() {
	pub.dict.Free()
}

// Free releases memory allocated to the private context's internal
// representation.
func (priv *PrivStore) Free() {
	priv.dict.Free()
}

// NewPubStoreFromProto creates a public store from its protobuf representation.
//
// You must call pub.Free() before pub goes out of scope.
func NewPubStoreFromProto(table *pb.Store) (pub *PubStore) {
	pub = new(PubStore)
	pub.dict = NewPubDictFromProto(table.GetDict())
	pub.sealed = table.GetSealed()
	pub.g = make(graph, table.GetNodeCt())
	for i := 0; i < len(table.Node); i++ {
		pub.g[table.Node[i]] = table.AdjList[i].Edge
	}
	return pub
}

// GetProto creates a protobuf representation of the public store.
//
// This is a compact representation suitable for transmission.
func (pub *PubStore) GetProto() *pb.Store {
	adjList := make([]*pb.Store_AdjList, 0)
	node := make([]int32, 0)
	for i := 0; i < len(pub.g); i++ {
		if len(pub.g[i]) > 0 {
			node = append(node, int32(i))
			adjList = append(adjList,
				&pb.Store_AdjList{Edge: pub.g[i]})
		}
	}
	return &pb.Store{
		Dict:    pub.dict.GetProto(),
		Sealed:  pub.sealed,
		Node:    node,
		AdjList: adjList,
		NodeCt:  int32(len(pub.g)),
	}
}

// String returns a string representing the public storage.
func (pub *PubStore) String() string {
	str := pub.dict.String()
	for i := 0; i < len(pub.sealed); i++ {
		str += fmt.Sprintf("%d: %s\n", i, hex.EncodeToString(pub.sealed[i]))
	}
	return str
}

// NewPrivStore creates a new private store context from a key and parameters.
//
// You must call priv.Free() before priv goes out of scope.
func NewPrivStore(K []byte, params *pb.Params) (priv *PrivStore, err error) {
	priv = new(PrivStore)

	block, err := aes.NewCipher(K[:DictKeyBytes])
	if err != nil {
		return nil, err
	}
	priv.aead, err = cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	priv.dict, err = NewPrivDict(K[DictKeyBytes:], params)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// GetParams returns the public parameters of the store.
func (priv *PrivStore) GetParams() *pb.Params {
	return priv.dict.GetParams()
}

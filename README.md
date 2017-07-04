Good times with data structures
===============================

A super light-weight password manager for paranoid people who like Go.

Package `store`
---------------

Package store provides secure storage of `map[string]string` objects. The
contents of the structure cannot be deduced from its public representation, and
querying it requires knowledge of a secret key. It is suitable for client/server
protocols whereby the service is trusted to provide storage, but is otherwise
not trusted.

The client possesses a secret key `K` and data `M` (of type `map[string]string`.
It executes:

```pub, priv, err := store.New(K, M)```

and transmits `pub`, the public representation of `M`, to the server.
To compute `M[input]`, the client executes:

```x, y, err := priv.GetIdx(input)```

and sends `x` and `y` to the server. These are integers corresponding to rows in
a table (of type `[][]byte`) encoded by `pub`. The server executes:

```pubShare, err := pub.GetShare(x, y)```

and sends `pubShare` (of type `[]byte`) to the client. (This is the XOR of the
two rows corresponding to `x` and `y`.) Finally, the client executes:

```output, err := priv.GetValue(input, pubShare)```

This combines `pubShare` with a private share computed from `input` and the key.
The result is `output = M[input]`.  Note that the server is not entrusted with
the key; its only job is to look up the rows of the table requested by the
client. The data structure is designed so that _no_ information about `input` or
`output` is leaked to any party not in possession of the secret key.

Note that the length of each `output` is limited to 60 bytes; see the Go
documentation for details.

Installation
------------

The core data structures are implemented in C. (See `c/`.) These are compiled
into a share object for which the Go code has bindings. They depend on OpenSSL,
so you'll need to install this library in advance. On Ubuntu:

```$ sudo apt-get install libssl-dev```

On Mac:

```$ brew install openssl```

To build the C code and run tests, go to `c/` and do

```$ make && make test```

Note that the tests will produce warnings from time to time. (This is OK as long
ast it doesn't produce **a lot** of warnings.) To install, do

```$ sudo make install && sudo ldconfig```

This builds a file called `libstruct.so` and moves it to `/usr/local/lib` and
copies the header files to `/usr/local/include/struct`.

Finally, you'll need Go. To get the latest version on Ubuntu, do

```
$ sudo add-apt-repository ppa:longsleep/golang-backports
$ sudo apt-get update
$ sudo apt-get install golang-go
```

On Mac, download the [pkg](https://golang.org/dl/) and install it.


Building `store.pb.go`
----------------------

**You only need to do this if you want to modify the code.** This project uses
protcool buffers and remote procedure calls. To build you'll first need the
lastest version of `protoc'. Go to [the gRPC
documentation](https://developers.google.com/protocol-buffers/docs/gotutorial)
for instructions. To build, run

```
  $ protoc -I . store.proto --go_out=plugins=grpc:.
```

in the root project directory.

Note Copyright
---------

This software is distributed under the terms of the 3-Clause BSD License; see
`LICENSE` in the root project directory for details.

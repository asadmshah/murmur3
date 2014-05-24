# Murmur3
Go package for the Murmur3 hash function by Austin Appleby referenced from
[here](https://code.google.com/p/smhasher/source/browse/trunk/MurmurHash3.cpp).
It's much faster than the standard library's hashing functions but slightly
slower than my xxhash package at https://github.com/asadmshah/xxhash.

### Installation
~~~
go get github.com/asadmshah/murmur3
~~~

### Usage
`New32` returns the standard library's `hash.Hash32` interface. `New128` returns
a `hash.Hash` interface with a `Sum128` method that returns 2 uint64s.
~~~
var seed uint64 = 1
h := murmur3.New128(seed)
h.Write([]byte("a"))
h.Sum128() // Returns 3983661951559065863, 15198864964193374602
~~~

Use `Checksum32` or `Checksum128` to quickly hash data.
~~~
var seed uint64 = 1
murmur3.Checksum128([]byte("a"), seed) // Returns 3983661951559065863, 15198864964193374602
~~~

### Benchmarks
Benchmarks are run using 1024 bytes.
~~~
BenchmarkChecksum32	    5000000	           499 ns/op	2050.58 MB/s	       0 B/op	       0 allocs/op
BenchmarkChecksum128	10000000	       233 ns/op	4386.30 MB/s	       0 B/op	       0 allocs/op
BenchmarkHashing32	    5000000	           669 ns/op	1529.28 MB/s	      16 B/op	       2 allocs/op
BenchmarkHashing128	    5000000	           446 ns/op	2292.30 MB/s	      32 B/op	       2 allocs/op
BenchmarkFNV	        1000000	           1411 ns/op	 725.62 MB/s	       8 B/op	       1 allocs/op
~~~


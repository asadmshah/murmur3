package murmur3

import (
	"encoding/binary"
	"hash"
	"unsafe"
)

const (
	c1_128 = 0x87c37b91114253d5
	c2_128 = 0x4cf5ad432745937f
)

// Hash128 is an interface for the 128-bit hash function.
type Hash128 interface {
	hash.Hash
	Sum128() (uint64, uint64)
}

type hash128 struct {
	size  int
	seed  uint64
	hash1 uint64
	hash2 uint64
	over  int
	tail  []byte
}

// New128 returns a new 128-bit hash.Hash initialized with seed.
func New128(seed uint64) Hash128 {
	h := new(hash128)
	h.seed = seed
	h.Reset()
	return h
}

func (self *hash128) BlockSize() int {
	return 16
}

func (self *hash128) Reset() {
	self.size = 0
	self.hash1 = self.seed
	self.hash2 = self.seed
	self.over = 0
	self.tail = make([]byte, 16)
}

func (self *hash128) Size() int {
	return 16
}

func (self *hash128) Sum(out []byte) []byte {
	b := make([]byte, 16)
	h1, h2 := self.Sum128()
	binary.BigEndian.PutUint64(b[:8], h1)
	binary.BigEndian.PutUint64(b[8:], h2)
	return append(out, b...)
}

func (self *hash128) Sum128() (uint64, uint64) {
	h1, h2 := tail128(self.tail[:self.over], self.hash1, self.hash2)
	return finalize128(h1, h2, self.size)
}

func (self *hash128) Write(data []byte) (int, error) {
	i := 0
	n := len(data)
	self.size += n

	// A blocksize of 16 is required in order to process.
	if n+self.over < 16 {
		copy(self.tail[self.over:], data)
		self.over += n
		return n, nil
	}

	// Write full blocksize of previous and current data.
	copy(self.tail[self.over:], data)
	i = 16 - self.over
	self.hash1, self.hash2, _ = body128(self.tail, 0, 16, self.hash1, self.hash2)
	self.over = 0

	// Write remaining data n/4 uint32 bytes of data.
	self.hash1, self.hash2, i = body128(data, i, n, self.hash1, self.hash2)

	// Place any remaining data in tail.
	copy(self.tail, data[i:])
	self.over = n - i

	return n, nil
}

// Checksum128 returns the 128-bit Murmur3 hash of data.
func Checksum128(data []byte, s uint64) (uint64, uint64) {
	n := len(data)
	h1, h2, i := body128(data, 0, n, s, s)
	h1, h2 = tail128(data[i:], h1, h2)
	h1, h2 = finalize128(h1, h2, n)
	return h1, h2
}

func body128(data []byte, i, n int, h1, h2 uint64) (uint64, uint64, int) {
	var k1, k2 uint64
	for ; i+16 <= n; i += 16 {
		k1 = *(*uint64)(unsafe.Pointer(&data[i]))
		k1 *= c1_128
		k1 = (k1 << 31) | (k1 >> 33)
		k1 *= c2_128

		h1 ^= k1
		h1 = (h1 << 27) | (h1 >> 37)
		h1 += h2
		h1 = h1*5 + 0x52dce729

		k2 = *(*uint64)(unsafe.Pointer(&data[i+8]))
		k2 *= c2_128
		k2 = (k2 << 33) | (k2 >> 31)
		k2 *= c1_128

		h2 ^= k2
		h2 = (h2 << 31) | (h2 >> 33)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}
	return h1, h2, i
}

func tail128(data []byte, h1, h2 uint64) (uint64, uint64) {
	var k1, k2 uint64

	i := len(data) & 15

	if i > 8 {
		for ; i > 8; i-- {
			k2 ^= uint64(data[i-1]) << uint((i-9)*8)
		}
		k2 *= c2_128
		k2 = (k2 << 33) | (k2 >> 31)
		k2 *= c1_128
		h2 ^= k2
	}

	if i > 0 {
		for ; i > 0; i-- {
			k1 ^= uint64(data[i-1]) << uint((i-1)*8)
		}
		k1 *= c1_128
		k1 = (k1 << 31) | (k1 >> 33)
		k1 *= c2_128
		h1 ^= k1
	}

	return h1, h2
}

func finalize128(h1, h2 uint64, n int) (uint64, uint64) {
	h1 ^= uint64(n)
	h2 ^= uint64(n)

	h1 += h2
	h2 += h1

	h1 = fmix64(h1)
	h2 = fmix64(h2)

	h1 += h2
	h2 += h1

	h1 = (h1 << 32) | (h1 >> 32)
	h2 = (h2 << 32) | (h2 >> 32)

	return h1, h2
}

func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

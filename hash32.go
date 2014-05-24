package murmur3

import (
	"encoding/binary"
	"hash"
	"unsafe"
)

const (
	c1_32 = 0xcc9e2d51
	c2_32 = 0x1b873593
)

type hash32 struct {
	size int
	seed uint32
	hash uint32
	over int
	tail []byte
}

// New32 returns a new 32-bit hash.Hash initialized with seed.
func New32(seed uint32) hash.Hash32 {
	h := new(hash32)
	h.seed = seed
	h.Reset()
	return h
}

func (self *hash32) BlockSize() int {
	return 4
}

func (self *hash32) Reset() {
	self.size = 0
	self.hash = self.seed
	self.over = 0
	self.tail = make([]byte, 4)
}

func (self *hash32) Size() int {
	return 4
}

func (self *hash32) Sum(out []byte) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, self.Sum32())
	return append(out, b...)
}

func (self *hash32) Sum32() uint32 {
	h := tail32(self.tail[:self.over], self.hash)
	return finalize32(h, self.size)
}

func (self *hash32) Write(data []byte) (int, error) {
	i := 0
	n := len(data)
	self.size += n

	// A blocksize of 4 is required in order to process.
	if n+self.over < 4 {
		copy(self.tail[self.over:], data)
		self.over += n
		return n, nil
	}

	// Write full blocksize of previous and current data.
	copy(self.tail[self.over:], data)
	i = 4 - self.over
	self.hash, _ = body32(self.tail, 0, 4, self.hash)
	self.over = 0

	// Write remaining data n/4 uint32 bytes of data.
	self.hash, i = body32(data, i, n, self.hash)

	// Place any remaining data in tail.
	copy(self.tail, data[i:])
	self.over = n - i

	return n, nil
}

// Checksum32 returns the 32-bit Murmur3 hash of data.
func Checksum32(data []byte, s uint32) uint32 {
	n := len(data)
	h, i := body32(data, 0, n, s)
	h = tail32(data[i:], h)
	h = finalize32(h, n)
	return h
}

func body32(data []byte, i, n int, h uint32) (uint32, int) {
	for k := uint32(0); i+4 <= n; i += 4 {
		k = *(*uint32)(unsafe.Pointer(&data[i]))
		k *= c1_32
		k = (k << 15) | (k >> 17)
		k *= c2_32

		h ^= k
		h = (h << 13) | (h >> 19)
		h = h*5 + 0xe6546b64
	}
	return h, i
}

func tail32(data []byte, h uint32) uint32 {
	k := uint32(0)
	switch len(data) & 3 {
	case 3:
		k ^= uint32(data[2]) << 16
		fallthrough
	case 2:
		k ^= uint32(data[1]) << 8
		fallthrough
	case 1:
		k ^= uint32(data[0])
		k *= c1_32
		k = (k << 15) | (k >> 17)
		k *= c2_32
		h ^= k
	}
	return h
}

func finalize32(h uint32, n int) uint32 {
	h ^= uint32(n)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

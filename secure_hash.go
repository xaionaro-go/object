package object

import (
	"crypto/sha512"
	"fmt"
	"hash"

	"lukechampine.com/blake3"
)

const blake3Size = 64 // 512 bits

type secureHash struct {
	Blake3 *blake3.Hasher
	SHA512 hash.Hash
}

var _ hash.Hash = (*secureHash)(nil)

func newSecureHash() *secureHash {
	return &secureHash{
		Blake3: blake3.New(blake3Size, nil),
		SHA512: sha512.New(),
	}
}

func (h *secureHash) Write(
	p []byte,
) (n int, err error) {
	n0, err := h.Blake3.Write(p)
	if err != nil {
		return n0, fmt.Errorf("unable to update the Blake3 part of the hash: %w", err)
	}

	n1, err := h.SHA512.Write(p)
	if err != nil {
		return max(n0, n1), fmt.Errorf("unable to update the SHA512 part of the hash: %w", err)
	}

	return max(n0, n1), nil
}

func (h *secureHash) Sum(
	b []byte,
) []byte {
	result := make([]byte, len(b)+h.Size())
	idx := 0

	copy(result[idx:], b)
	idx += len(b)

	h0 := h.Blake3.Sum(nil)
	copy(result[idx:], h0)
	idx += len(h0)

	h1 := h.SHA512.Sum(nil)
	copy(result[idx:], h1)
	idx += len(h1)

	return result
}

func (h *secureHash) Reset() {
	h.Blake3.Reset()
	h.SHA512.Reset()
}

func (h *secureHash) Size() int {
	return h.Blake3.Size() + h.SHA512.Size()
}

func (h *secureHash) BlockSize() int {
	return max(h.Blake3.BlockSize(), h.SHA512.BlockSize())
}

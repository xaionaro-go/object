package object

import (
	"bytes"
)

// Hash is a hash of an value/object.
type Hash []byte

// Equals returns true if the hash is equal to the provided hash
func (h Hash) Equals(b Hash) bool {
	return bytes.Equal(h, b)
}

// Less returns true if the hash is ordered earlier than the provided hash
// (could be useful for implementing sort.Interface).
func (h Hash) Less(b Hash) bool {
	return bytes.Compare(h, b) < 0
}

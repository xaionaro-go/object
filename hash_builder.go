package object

import (
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/xaionaro-go/unsafetools"
)

// HashBuilder is the handler which converts a set of variables to a Hash.
type HashBuilder struct {
	locker        sync.Mutex
	buffer        [16]byte
	HashValue     hash.Hash
	StableHashing bool
	byteOrder     binary.ByteOrder
}

// NewBuilderUnstable returns a new instance of HashBuilder that
// builds unstable hashes (that change for the same object
// after each restart of the program).
func NewHashBuilderUnstable(hash hash.Hash) *HashBuilder {
	return &HashBuilder{
		HashValue:     hash,
		StableHashing: false,
		byteOrder:     binary.NativeEndian,
	}
}

// NewBuilderStable returns a new instance of HashBuilder that
// builds stable hashes (that do not change for the same object
// after each restart of the program).
func NewHashBuilderStable(hash hash.Hash) *HashBuilder {
	return &HashBuilder{
		HashValue:     hash,
		StableHashing: true,
		byteOrder:     binary.LittleEndian,
	}
}

func (b *HashBuilder) extend(in []byte) error {
	h := b.HashValue

	oldHash := h.Sum(nil)

	_, err := h.Write(oldHash)
	if err != nil {
		return fmt.Errorf("unable to write old hash: %w", err)
	}

	_, err = h.Write(in)
	if err != nil {
		return fmt.Errorf("unable to extend %T: %w", h, err)
	}

	return nil
}

// Write adds more measurements to the current hash.
func (b *HashBuilder) Write(args ...any) error {
	b.locker.Lock()
	defer b.locker.Unlock()
	return b.write(args...)
}

func (b *HashBuilder) write(args ...any) error {
	var writeType func(t reflect.Type) error
	if b.StableHashing {
		writeType = func(t reflect.Type) error {
			return b.writeString(t.PkgPath(), ".", t.Name())
		}
	} else {
		writeType = func(t reflect.Type) error {
			typePtr := reflect.ValueOf(t).Pointer()
			return b.writeUintptr(typePtr)
		}
	}

	for idx, obj := range args {
		err := Traverse(
			obj,
			func(
				ctx *ProcContext,
				v reflect.Value,
				sf *reflect.StructField,
			) (reflect.Value, bool, error) {
				t := v.Type()
				if err := writeType(t); err != nil {
					return v, false, fmt.Errorf("unable to extend the type '%s' pointer of the object: %w", t, err)
				}
				shouldContinue, err := b.writeValue(v)
				if err != nil {
					return v, false, fmt.Errorf("unable to extend the value of type '%s': %w", t, err)
				}
				return v, shouldContinue, nil
			},
		)
		if err != nil {
			return fmt.Errorf("unable to traverse&hash argument #%d of type %T: %w", idx, obj, err)
		}
	}
	return nil
}

func (b *HashBuilder) getBuffer(size uint) []byte {
	return b.buffer[:size]
}

func (b *HashBuilder) writeString(ss ...string) error {
	h := b.HashValue

	oldHash := h.Sum(nil)

	_, err := h.Write(oldHash)
	if err != nil {
		return fmt.Errorf("unable to write old hash: %w", err)
	}

	for _, s := range ss {
		_, err = h.Write(unsafetools.CastStringToBytes(s))
		if err != nil {
			return fmt.Errorf("unable to extend string '%s': %w", s, err)
		}
	}

	return nil
}

func (b *HashBuilder) writeValue(v reflect.Value) (bool, error) {
	switch v.Kind() {
	case reflect.Bool:
		return false, b.writeBool(v.Bool())
	case reflect.Int:
		return false, b.writeUint(uint(v.Int()))
	case reflect.Int8:
		return false, b.writeUint8(uint8(v.Int()))
	case reflect.Int16:
		return false, b.writeUint16(uint16(v.Int()))
	case reflect.Int32:
		return false, b.writeUint32(uint32(v.Int()))
	case reflect.Int64:
		return false, b.writeUint64(uint64(v.Int()))
	case reflect.Uint:
		return false, b.writeUint(uint(v.Uint()))
	case reflect.Uint8:
		return false, b.writeUint8(uint8(v.Uint()))
	case reflect.Uint16:
		return false, b.writeUint16(uint16(v.Uint()))
	case reflect.Uint32:
		return false, b.writeUint32(uint32(v.Uint()))
	case reflect.Uint64:
		return false, b.writeUint64(uint64(v.Uint()))
	case reflect.Uintptr:
		return false, b.writeUintptr(uintptr(v.Uint()))
	case reflect.Float32:
		return false, b.writeFloat32(float32(v.Float()))
	case reflect.Float64:
		return false, b.writeFloat64(float64(v.Float()))
	case reflect.Complex64:
		return false, b.writeComplex64(complex64(v.Complex()))
	case reflect.Complex128:
		return false, b.writeComplex128(v.Complex())
	case reflect.Array:
		// the items of the array will be traversed by Traverse, we need to write the length only here
		return true, b.writeUint(uint(v.Len()))
	case reflect.Chan:
		return false, fmt.Errorf("unable to serialize a channel")
	case reflect.Func:
		return false, fmt.Errorf("unable to serialize a function")
	case reflect.Interface:
		return false, nil
	case reflect.Map:
		return false, b.writeMapRF(v)
	case reflect.Pointer:
		// asking to traverse it:
		return true, nil
	case reflect.Slice:
		// the items of the slice will be traversed by Traverse, we need to write the length only here
		return true, b.writeUint(uint(v.Len()))
	case reflect.String:
		return false, b.writeString(v.String())
	case reflect.Struct:
		return true, nil
	case reflect.UnsafePointer:
		return false, fmt.Errorf("cannot serialize an unsafe pointer")
	default:
		return false, fmt.Errorf("unexpected kind: %v", v.Kind())
	}
}

func (b *HashBuilder) writeUintptr(v uintptr) error {
	size := uintptrSize
	buf := b.getBuffer(size)
	switch size {
	case 4:
		b.byteOrder.PutUint32(buf, uint32(v))
	case 8:
		b.byteOrder.PutUint64(buf, uint64(v))
	default:
		return fmt.Errorf("unexpected size of uintptr: %d", size)
	}
	return b.extend(buf)
}

// Reset returns current hash.
func (b *HashBuilder) Result() []byte {
	b.locker.Lock()
	defer b.locker.Unlock()
	return b.result()
}
func (b *HashBuilder) result() []byte {
	return b.HashValue.Sum(nil)
}

// Reset resets the state of the hash.
func (b *HashBuilder) Reset() {
	b.locker.Lock()
	defer b.locker.Unlock()
	b.reset()
}

func (b *HashBuilder) reset() {
	b.HashValue.Reset()
}

// ResetAndHash resets the state of the hash, calculates a new hash
// using given arguments, and returns it's value.
func (b *HashBuilder) ResetAndHash(args ...any) (Hash, error) {
	b.locker.Lock()
	defer b.locker.Unlock()
	return b.resetAndHash(args...)
}

func (b *HashBuilder) resetAndHash(args ...any) (Hash, error) {
	b.reset()
	err := b.write(args...)
	if err != nil {
		return nil, fmt.Errorf("unable to write the values to ")
	}
	return b.result(), nil
}

func (b *HashBuilder) writeBool(v bool) error {
	u := uint8(0)
	if v {
		u = 1
	}
	return b.writeUint8(u)
}
func (b *HashBuilder) writeUint(v uint) error {
	size := uintSize
	buf := b.getBuffer(size)
	switch size {
	case 4:
		b.byteOrder.PutUint32(buf, uint32(v))
	case 8:
		b.byteOrder.PutUint64(buf, uint64(v))
	default:
		return fmt.Errorf("unexpected size of int: %d", size)
	}
	return b.extend(buf)
}
func (b *HashBuilder) writeUint8(v uint8) error {
	b.buffer[0] = v
	return b.extend(b.buffer[:1])
}
func (b *HashBuilder) writeUint16(v uint16) error {
	b.byteOrder.PutUint16(b.buffer[:], v)
	return b.extend(b.buffer[:2])
}
func (b *HashBuilder) writeUint32(v uint32) error {
	b.byteOrder.PutUint32(b.buffer[:], v)
	return b.extend(b.buffer[:4])
}
func (b *HashBuilder) writeUint64(v uint64) error {
	b.byteOrder.PutUint64(b.buffer[:], v)
	return b.extend(b.buffer[:8])
}
func (b *HashBuilder) writeFloat32(v float32) error {
	b.byteOrder.PutUint32(b.buffer[:], math.Float32bits(v))
	return b.extend(b.buffer[:4])
}
func (b *HashBuilder) writeFloat64(v float64) error {
	b.byteOrder.PutUint64(b.buffer[:], math.Float64bits(v))
	return b.extend(b.buffer[:8])
}
func (b *HashBuilder) writeComplex64(v complex64) error {
	b.byteOrder.PutUint32(b.buffer[0:], math.Float32bits(real(v)))
	b.byteOrder.PutUint32(b.buffer[4:], math.Float32bits(imag(v)))
	return b.extend(b.buffer[:8])
}
func (b *HashBuilder) writeComplex128(v complex128) error {
	b.byteOrder.PutUint64(b.buffer[0:], math.Float64bits(real(v)))
	b.byteOrder.PutUint64(b.buffer[8:], math.Float64bits(imag(v)))
	return b.extend(b.buffer[:16])
}

type valuesAndHashes struct {
	Values []reflect.Value
	Hashes []Hash
}

var _ sort.Interface = (*valuesAndHashes)(nil)

func (s *valuesAndHashes) Len() int {
	return len(s.Values)
}
func (s *valuesAndHashes) Less(i, j int) bool {
	return s.Hashes[i].Less(s.Hashes[j])
}
func (s *valuesAndHashes) Swap(i, j int) {
	s.Values[i], s.Hashes[i], s.Values[j], s.Hashes[j] = s.Values[j], s.Hashes[j], s.Values[i], s.Hashes[i]
}

func (b *HashBuilder) writeMapRF(v reflect.Value) error {
	keys := v.MapKeys()
	hashes := make([]Hash, len(keys))
	subBuilder := NewHashBuilderStable(newSecureHash())
	for idx, key := range keys {
		var err error
		hashes[idx], err = subBuilder.resetAndHash(key)
		if err != nil {
			return fmt.Errorf("unable to hash map key of type %s: %w", key.Type(), err)
		}
	}
	s := &valuesAndHashes{
		Values: keys,
		Hashes: hashes,
	}
	sort.Sort(s)
	for _, key := range s.Values {
		err := b.write(key)
		if err != nil {
			return fmt.Errorf("unable to write map key of type '%s': %w", key.Type(), err)
		}
		mapValue := v.MapIndex(key)
		err = b.write(mapValue)
		if err != nil {
			return fmt.Errorf("unable to write map value of type '%s': %w", mapValue.Type(), err)
		}
	}
	return nil
}

var defaultHashBuilder = NewHashBuilderStable(newSecureHash())

// CalcCryptoHash returns a cryptographically secure hash of an arbitrary set of values.
func CalcCryptoHash(args ...any) (Hash, error) {
	return defaultHashBuilder.ResetAndHash(args...)
}

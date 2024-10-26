package object

import (
	"unsafe"
)

var (
	uintSize    uint
	uintptrSize uint
)

func init() {
	uintSize = uint(unsafe.Sizeof(uint(0)))
	uintptrSize = uint(unsafe.Sizeof(uintptr(0)))
}

package object

import (
	"fmt"
	"reflect"
)

// Traverse recursively traverses the object `obj`.
func Traverse(
	obj any,
	visitorFunc VisitorFunc,
) error {
	_, err := newTraverser().traverse(reflect.ValueOf(obj), visitorFunc, newProcContext(), nil)
	return err
}

// VisitorFunc is called on every node during a traversal.
type VisitorFunc func(*ProcContext, reflect.Value, *reflect.StructField) (reflect.Value, bool, error)

// ProcContext is a structure provided to a callback on every call.
type ProcContext struct {
	parent *ProcContext
	path   string
	depth  uint

	// CustomData is overwritable and all the children in the tree
	// will receive this provided value.
	CustomData any
}

func (ctx *ProcContext) Parent() *ProcContext {
	return ctx.parent
}

func (ctx *ProcContext) Path() string {
	return ctx.path
}

func (ctx *ProcContext) Depth() uint {
	return ctx.depth
}

func newProcContext() *ProcContext {
	return &ProcContext{}
}

func (ctx *ProcContext) Next(pathPart string) *ProcContext {
	return &ProcContext{
		parent:     ctx,
		path:       ctx.path + "." + pathPart,
		depth:      ctx.depth + 1,
		CustomData: ctx.CustomData,
	}
}

type traverser struct {
	AlreadyVisitedPointers map[uintptr]struct{}
}

func newTraverser() *traverser {
	return &traverser{}
}

func (traverser *traverser) traverse(
	v reflect.Value,
	visitorFunc VisitorFunc,
	ctx *ProcContext,
	structField *reflect.StructField,
) (_ret reflect.Value, _err error) {
	newV, goInside, err := visitorFunc(ctx, v, structField)
	if err != nil {
		return newV, fmt.Errorf("received an error from the visitor function at '%s': %w", ctx.path, err)
	}
	if !goInside {
		return newV, nil
	}
	v = newV

	t := v.Type()

	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			idxV := v.Index(i)
			newV, err := traverser.traverse(v.Index(i), visitorFunc, ctx.Next(fmt.Sprintf("[%d]", i)), nil)
			if newV != idxV {
				idxV.Set(newV)
			}
			if err != nil {
				return v, err
			}
		}
	case reflect.Interface:
		if !v.Elem().IsValid() {
			return v, nil
		}
		newV, err := traverser.traverse(v.Elem(), visitorFunc, ctx.Next("{}"), nil)
		if newV != v.Elem() {
			v.Set(newV)
		}
		if err != nil {
			return v, err
		}
	case reflect.Map:
		if v.IsNil() {
			return v, nil
		}
		iter := v.MapRange()
		for iter.Next() {
			mapK := iter.Key()
			mapV := iter.Value()
			newV, err := traverser.traverse(mapV, visitorFunc, ctx.Next(fmt.Sprintf("[%v]", mapK)), nil)
			if newV != mapV {
				v.SetMapIndex(mapK, newV)
			}
			if err != nil {
				return v, err
			}
		}
	case reflect.Pointer:
		if v.IsNil() {
			return v, nil
		}
		ptr := v.Pointer()
		if traverser.AlreadyVisitedPointers == nil {
			traverser.AlreadyVisitedPointers = make(map[uintptr]struct{})
		}
		if _, ok := traverser.AlreadyVisitedPointers[ptr]; ok {
			return v, nil
		}
		traverser.AlreadyVisitedPointers[ptr] = struct{}{}
		vElem := v.Elem()
		newV, err := traverser.traverse(vElem, visitorFunc, ctx.Next("*"), nil)
		if newV != vElem {
			vElem.Set(newV)
		}
		if err != nil {
			return v, err
		}
	case reflect.Slice:
		if v.IsNil() {
			return v, nil
		}
		for i := 0; i < v.Len(); i++ {
			idxV := v.Index(i)
			newV, err := traverser.traverse(idxV, visitorFunc, ctx.Next(fmt.Sprintf("[%d]", i)), nil)
			if newV != idxV {
				idxV.Set(newV)
			}
			if err != nil {
				return v, err
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fV := v.Field(i)
			fT := t.Field(i)

			if fT.PkgPath != "" {
				// unexported
				continue
			}

			newV, err := traverser.traverse(fV, visitorFunc, ctx.Next(fT.Name), &fT)
			if newV != fV {
				if !fV.CanSet() {
					newStruct := reflect.New(v.Type()).Elem()
					newStruct.Set(v)
					v = newStruct
					fV = v.Field(i)
				}
				fV.Set(newV)
			}
			if err != nil {
				return v, err
			}
		}
	}
	return v, nil
}

// Pointer is a constraint for pointers only.
type Pointer[T any] interface {
	*T
}

// RemoveSecrets returns zero-s all fields tagged as `secret:""`.
//
// Keep in mind, this function does not zero:
// * the internals of: channels, function values, uintptr-s and unsafe.Pointer-s;
// * the keys of maps.
//
// Also, it does not copy unexported data!
func RemoveSecrets[T any, PTR Pointer[T]](obj PTR) {
	type markerIsSecretT struct{}
	var markerIsSecret markerIsSecretT
	err := Traverse(obj, func(ctx *ProcContext, v reflect.Value, sf *reflect.StructField) (reflect.Value, bool, error) {
		if sf == nil {
			return v, true, nil
		}

		_, isSecret := sf.Tag.Lookup("secret")
		if !isSecret {
			isSecret = ctx.CustomData == markerIsSecret
		}
		if !isSecret {
			return v, true, nil
		}
		ctx.CustomData = markerIsSecret

		return reflect.Zero(v.Type()), false, nil
	})
	if err != nil {
		panic(err)
	}
}

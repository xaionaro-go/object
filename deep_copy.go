package object

import (
	"fmt"
	"reflect"

	"github.com/xaionaro-go/unsafetools"
)

// DeepCopy returns a deep copy of the object.
//
// Keep in mind, by default it does not copy unexported data (unless
// option `WithUnexported(true)` is provided).
func DeepCopy[T any](
	obj T,
	opts ...Option,
) T {
	cfg := Options(opts).config()
	v := reflect.ValueOf(&obj)
	return *newDeepCopier(cfg).deepCopy(v, newProcContext(), nil).Interface().(*T)
}

type ProcFunc func(*ProcContext, reflect.Value, *reflect.StructField) reflect.Value

type deepCopier struct {
	config                     config
	copiedValuesBehindPointers map[uintptr]reflect.Value
}

func newDeepCopier(cfg config) *deepCopier {
	return &deepCopier{
		config: cfg,
	}
}

func (c *deepCopier) deepCopy(
	v reflect.Value,
	ctx *ProcContext,
	structField *reflect.StructField,
) (_ret reflect.Value) {
	if c.config.ProcFunc != nil {
		defer func() {
			_ret = c.config.ProcFunc(ctx, _ret, structField)
		}()
	}

	t := v.Type()
	result := reflect.New(t).Elem()

	switch v.Kind() {
	case reflect.Bool:
		result.Set(v)
	case reflect.Int:
		result.Set(v)
	case reflect.Int8:
		result.Set(v)
	case reflect.Int16:
		result.Set(v)
	case reflect.Int32:
		result.Set(v)
	case reflect.Int64:
		result.Set(v)
	case reflect.Uint:
		result.Set(v)
	case reflect.Uint8:
		result.Set(v)
	case reflect.Uint16:
		result.Set(v)
	case reflect.Uint32:
		result.Set(v)
	case reflect.Uint64:
		result.Set(v)
	case reflect.Uintptr, reflect.UnsafePointer:
		// We assume that if somebody uses uintptr or/and unsafe.Pointer
		// then they take all the responsibility for whatever happens,
		// so we just copy as is.
		result.Set(v)
	case reflect.Float32:
		result.Set(v)
	case reflect.Float64:
		result.Set(v)
	case reflect.Complex64:
		result.Set(v)
	case reflect.Complex128:
		result.Set(v)
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			result.Index(i).Set(c.deepCopy(v.Index(i), ctx.Next(fmt.Sprintf("[%d]", i)), nil))
		}
	case reflect.Chan:
		result.Set(v)
	case reflect.Func:
		result.Set(v)
	case reflect.Interface:
		if !v.Elem().IsValid() { // if unwrapInterface(v) == nil { return v }
			return v
		}
		result.Set(c.deepCopy(v.Elem(), ctx.Next("{}"), nil))
	case reflect.Map:
		if v.IsNil() {
			return result
		}
		result = reflect.MakeMapWithSize(t, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			result.SetMapIndex(k, c.deepCopy(v, ctx.Next(fmt.Sprintf("[%v]", k)), nil))
		}
	case reflect.Pointer:
		if v.IsNil() {
			return result
		}
		ptr := v.Pointer()
		if c.copiedValuesBehindPointers == nil {
			c.copiedValuesBehindPointers = make(map[uintptr]reflect.Value)
		}
		if v, ok := c.copiedValuesBehindPointers[ptr]; ok {
			return v
		}
		c.copiedValuesBehindPointers[ptr] = result
		result.Set(reflect.New(t.Elem()))                           // result = &T{}
		result.Elem().Set(c.deepCopy(v.Elem(), ctx.Next("*"), nil)) // *result = *v
	case reflect.Slice:
		if v.IsNil() {
			return result
		}
		result = reflect.MakeSlice(t, v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			result.Index(i).Set(c.deepCopy(v.Index(i), ctx.Next(fmt.Sprintf("[%d]", i)), nil))
		}
	case reflect.String:
		result.Set(v)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fV := v.Field(i)
			fT := t.Field(i)

			if fT.PkgPath != "" {
				if !c.config.ProcessUnexported {
					// unexported
					continue
				}
				if !v.CanAddr() {
					newV := reflect.New(v.Type()).Elem()
					newV.Set(v)
					v = newV
				}
				fV = unsafetools.FieldByIndexInValue(v.Addr(), i)
			}

			newV := c.deepCopy(fV, ctx.Next(fT.Name), &fT)
			if fT.PkgPath != "" {
				// unexported
				out := unsafetools.FieldByIndexInValue(result.Addr(), i)
				if out.CanSet() {
					out.Set(newV)
				} else {
					// TODO: investigate why CanSet could be false
					out.Elem().Set(newV.Elem())
				}
			} else {
				result.Field(i).Set(newV)
			}
		}
	default:
		panic(fmt.Errorf("unexpected kind: %v", v.Kind()))
	}
	return result
}

// DeepCopyWithoutSecrets returns a deep copy of the object, but with all
// fields tagged as `secret:""` reset to their zero values.
//
// Keep in mind, this function does not censor:
// * the internals of: channels, function values, uintptr-s and unsafe.Pointer-s;
// * the keys of maps.
//
// Also, it does not copy unexported data.
func DeepCopyWithoutSecrets[T any](
	obj T,
	opts ...Option,
) T {
	cfg := Options(opts).config()
	return DeepCopy(
		obj,
		OptionWithProcessingFunc(func(
			ctx *ProcContext,
			v reflect.Value,
			sf *reflect.StructField,
		) reflect.Value {
			if cfg.ProcFunc != nil {
				v = cfg.ProcFunc(ctx, v, sf)
			}

			if sf == nil {
				return v
			}

			if _, ok := sf.Tag.Lookup("secret"); !ok {
				return v
			}

			return reflect.Zero(v.Type())
		}),
		OptionWithUnexported(cfg.ProcessUnexported),
	)
}

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
	result, _, err := newDeepCopier(cfg).deepCopy(v, newProcContext(), nil)
	if err != nil {
		panic(err)
	}
	return *result.Interface().(*T)
}

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
) (reflect.Value, bool, error) {
	if c.config.VisitorFunc != nil {
		newV, goDeeper, err := c.config.VisitorFunc(ctx, v, structField)
		v = newV
		if err != nil {
			return v, goDeeper, fmt.Errorf("got an error from the visitor function at '%s': %w", ctx.path, err)
		}
		if !goDeeper {
			return v, false, nil
		}
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
			c, _, err := c.deepCopy(v.Index(i), ctx.Next(fmt.Sprintf("[%d]", i)), nil)
			if err != nil {
				return reflect.Value{}, false, err
			}
			result.Index(i).Set(c)
		}
	case reflect.Chan:
		result.Set(v)
	case reflect.Func:
		result.Set(v)
	case reflect.Interface:
		if !v.Elem().IsValid() { // if unwrapInterface(v) == nil { return v }
			return v, false, nil
		}
		newV, _, err := c.deepCopy(v.Elem(), ctx.Next("{}"), nil)
		if err != nil {
			return result, false, err
		}
		result.Set(newV)
	case reflect.Map:
		if v.IsNil() {
			return result, false, nil
		}
		result = reflect.MakeMapWithSize(t, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			newV, _, err := c.deepCopy(v, ctx.Next(fmt.Sprintf("[%v]", k)), nil)
			if err != nil {
				return result, false, err
			}
			result.SetMapIndex(k, newV)
		}
	case reflect.Pointer:
		if v.IsNil() {
			return result, false, nil
		}
		ptr := v.Pointer()
		if c.copiedValuesBehindPointers == nil {
			c.copiedValuesBehindPointers = make(map[uintptr]reflect.Value)
		}
		if v, ok := c.copiedValuesBehindPointers[ptr]; ok {
			return v, false, nil
		}
		c.copiedValuesBehindPointers[ptr] = result
		result.Set(reflect.New(t.Elem())) // result = &T{}
		newVElem, _, err := c.deepCopy(v.Elem(), ctx.Next("*"), nil)
		if err != nil {
			return result, false, err
		}
		result.Elem().Set(newVElem) // *result = *v
	case reflect.Slice:
		if v.IsNil() {
			return result, false, nil
		}
		result = reflect.MakeSlice(t, v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			newV, _, err := c.deepCopy(v.Index(i), ctx.Next(fmt.Sprintf("[%d]", i)), nil)
			if err != nil {
				return result, false, err
			}
			result.Index(i).Set(newV)
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
					vWithAddr := reflect.New(v.Type()).Elem()
					vWithAddr.Set(v)
					if v.Type() != vWithAddr.Type() {
						panic(fmt.Errorf("internal error: received wrong type at '%s': expected:%s, received:%s", ctx.path, v.Type(), vWithAddr.Type()))
					}
					v = vWithAddr
				}
				fVWithAddr := unsafetools.FieldByIndexInValue(v.Addr(), i).Elem()
				if fV.Type() != fVWithAddr.Type() {
					panic(fmt.Errorf("internal error: received wrong type at '%s': expected:%s, received:%s", ctx.path, fV.Type(), fVWithAddr.Type()))
				}
				fV = fVWithAddr
			}

			newFV, _, err := c.deepCopy(fV, ctx.Next(fT.Name), &fT)
			if err != nil {
				return result, false, err
			}
			outF := result.Field(i)
			if outF.Type() != newFV.Type() {
				panic(fmt.Errorf("received wrong type at '%s': expected:%s, received:%s", ctx.path, outF.Type(), newFV.Type()))
			}
			if fT.PkgPath != "" {
				// unexported
				outF = unsafetools.FieldByIndexInValue(result.Addr(), i).Elem()
			}
			outF.Set(newFV)
		}
	default:
		panic(fmt.Errorf("unexpected kind: %v", v.Kind()))
	}
	return result, false, nil
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
		OptionWithVisitorFunc(func(
			ctx *ProcContext,
			v reflect.Value,
			sf *reflect.StructField,
		) (reflect.Value, bool, error) {
			goDeeper := true
			if cfg.VisitorFunc != nil {
				var err error
				v, goDeeper, err = cfg.VisitorFunc(ctx, v, sf)
				if err != nil {
					return v, goDeeper, err
				}
			}

			if sf == nil {
				return v, goDeeper, nil
			}

			if _, ok := sf.Tag.Lookup("secret"); !ok {
				return v, goDeeper, nil
			}

			return reflect.Zero(v.Type()), goDeeper, nil
		}),
		OptionWithUnexported(cfg.ProcessUnexported),
	)
}

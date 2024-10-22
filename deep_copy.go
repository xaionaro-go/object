package deepcopy

import (
	"fmt"
	"reflect"
)

// DeepCopy returns a deep copy of the object.
//
// Keep in mind, it does not copy unexported data.
func DeepCopy[T any](obj T) T {
	v := reflect.ValueOf(obj)
	return deepCopy(v, nil, nil).Interface().(T)
}

type ProcFunc func(reflect.Value, *reflect.StructField) reflect.Value

// DeepCopyWithProcessing does the same thing as DeepCopy, but
// also provides possibility to modify the data during the copy.
//
// Keep in mind, this function does not process:
// * the internals of channels, function values, uintptr-s and unsafe.Pointer-s;
// * the keys of maps.
func DeepCopyWithProcessing[T any](
	obj T,
	procFunc ProcFunc,
) T {
	v := reflect.ValueOf(obj)
	return deepCopy(v, procFunc, nil).Interface().(T)
}

func deepCopy(
	v reflect.Value,
	procFunc ProcFunc,
	structField *reflect.StructField,
) (_ret reflect.Value) {
	if procFunc != nil {
		defer func() {
			_ret = procFunc(_ret, structField)
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
			result.Index(i).Set(deepCopy(v.Index(i), procFunc, nil))
		}
	case reflect.Chan:
		result.Set(v)
	case reflect.Func:
		result.Set(v)
	case reflect.Interface:
		if v.IsNil() {
			return result
		}
		result.Set(deepCopy(v.Elem(), procFunc, nil))
	case reflect.Map:
		if v.IsNil() {
			return result
		}
		result = reflect.MakeMapWithSize(t, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			result.SetMapIndex(k, deepCopy(v, procFunc, nil))
		}
	case reflect.Pointer:
		if v.IsNil() {
			return result
		}
		result.Set(reflect.New(t.Elem()))                    // result = (*T)(nil)
		result.Elem().Set(deepCopy(v.Elem(), procFunc, nil)) // *result = *v
	case reflect.Slice:
		if v.IsNil() {
			return result
		}
		result = reflect.MakeSlice(t, v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			result.Index(i).Set(deepCopy(v.Index(i), procFunc, nil))
		}
	case reflect.String:
		result.Set(v)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fV := v.Field(i)
			fT := t.Field(i)

			if fT.PkgPath != "" {
				// unexported
				continue
			}

			result.Field(i).Set(deepCopy(fV, procFunc, &fT))
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
// * the internals of channels, function values, uintptr-s and unsafe.Pointer-s;
// * the keys of maps.
//
// Also, it does not copy unexported data.
func DeepCopyWithoutSecrets[T any](obj T) T {
	return DeepCopyWithProcessing(obj, func(v reflect.Value, sf *reflect.StructField) reflect.Value {
		if sf == nil {
			return v
		}

		if _, ok := sf.Tag.Lookup("secret"); !ok {
			return v
		}

		return reflect.Zero(v.Type())
	})
}

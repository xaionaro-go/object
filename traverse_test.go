package object

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTraverseZerofy(t *testing.T) {
	sample := testSample()
	err := Traverse(
		sample,
		func(ctx *ProcContext, v reflect.Value, sf *reflect.StructField) (reflect.Value, bool, error) {
			if sf == nil {
				return v, true, nil
			}
			return reflect.Zero(v.Type()), false, nil
		},
	)
	require.NoError(t, err)

	sample.unexpectedField = ""
	require.Equal(t, &testType{}, sample)
}

func TestRemoveSecrets(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		sample := testSample()
		RemoveSecrets(sample)
		require.Equal(t, testSampleWithoutSecrets(), sample)
	})
	t.Run("iface-pointer", func(t *testing.T) {
		sample := testSample()
		var iface any = *sample
		RemoveSecrets(&iface)
		require.Equal(t, *testSampleWithoutSecrets(), iface)
	})
}

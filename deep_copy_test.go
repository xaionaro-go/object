package deepcopy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mapKey struct {
	SomePublic int
	SomeSecret int `secret:""`
}

type testType struct {
	SomeSlice        []testType
	SomeMap          map[mapKey]testType
	SomePublicString string
	SomeSecretString string `secret:""`
	SomePointer      *testType
	SomeError        error
}

func TestNilInterface(t *testing.T) {
	require.Equal(t, error(nil), DeepCopy(error(nil)))
}

func Test(t *testing.T) {
	sample := &testType{
		SomeSlice: nil,
		SomeMap: map[mapKey]testType{
			{1, 2}: {
				SomePublicString: "hello",
				SomeSecretString: "bye",
			},
		},
		SomePublicString: "public secret",
		SomeSecretString: "secret secret",
		SomePointer: &testType{
			SomeSlice: []testType{{
				SomePublicString: "false == false",
				SomeSecretString: "but sometimes it does not",
				SomePointer:      &testType{},
			}},
			SomePublicString: "true == true",
			SomeSecretString: "but there is a nuance",
		},
	}

	require.Equal(t, sample, DeepCopy(sample))
	require.Equal(t, &testType{
		SomeSlice: nil,
		SomeMap: map[mapKey]testType{
			{1, 2}: {
				SomePublicString: "hello",
			},
		},
		SomePublicString: "public secret",
		SomePointer: &testType{
			SomeSlice: []testType{{
				SomePublicString: "false == false",
				SomePointer:      &testType{},
			}},
			SomePublicString: "true == true",
		},
	}, DeepCopyWithoutSecrets(sample))
}

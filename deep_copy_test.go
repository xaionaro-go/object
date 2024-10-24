package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type mapKey struct {
	SomePublic int
	SomeSecret int `secret:""`
}

type testType struct {
	SomeSlice            []testType
	SomeMap              map[mapKey]testType
	SomePublicString     string
	SomeSecretString     string `secret:""`
	SomePointer          *testType
	SomeError            error
	SomeSecretSliceOfAny []any            `secret:""`
	SomeSecretMap        map[int]testType `secret:""`
	unexpectedField      string
}

func TestNilInterface(t *testing.T) {
	require.Equal(t, error(nil), DeepCopy(error(nil)))
}

func TestInfiniteRecursion(t *testing.T) {
	s := &testType{
		SomePublicString: "1",
	}
	s.SomePointer = s
	require.Equal(t, s, DeepCopy(s))
}

func testSample() *testType {
	return &testType{
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
		SomeSecretSliceOfAny: []any{"random secrets"},
		SomeSecretMap:        map[int]testType{1: {}},
		unexpectedField:      "unexpected data",
	}
}

func testSampleWithoutSecrets() *testType {
	return &testType{
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
		unexpectedField: "unexpected data",
	}
}

func TestComplexStructure(t *testing.T) {
	sample := testSample()
	sample.unexpectedField = ""
	require.Equal(t, sample, DeepCopy(sample))

	sampleWithoutSecrets := testSampleWithoutSecrets()
	sampleWithoutSecrets.unexpectedField = ""
	require.Equal(t, sampleWithoutSecrets, DeepCopyWithoutSecrets(sample))
}

func TestComplexStructureWithUnexported(t *testing.T) {
	sample := testSample()
	require.Equal(t, sample, DeepCopy(sample, OptionWithUnexported(true)))

	sampleWithoutSecrets := testSampleWithoutSecrets()
	require.Equal(t, sampleWithoutSecrets, DeepCopyWithoutSecrets(sample, OptionWithUnexported(true)))
}

package object

import (
	"testing"
	"time"

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
	SomeTime             time.Time
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
		SomeTime:             time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC),
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
		SomeTime:        time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC),
		unexpectedField: "unexpected data",
	}
}

func TestComplexStructure(t *testing.T) {
	sample := testSample()
	sample.SomeTime = time.Time{}
	sample.unexpectedField = ""
	require.Equal(t, sample, DeepCopy(sample))

	sampleWithoutSecrets := testSampleWithoutSecrets()
	sampleWithoutSecrets.SomeTime = time.Time{}
	sampleWithoutSecrets.unexpectedField = ""
	require.Equal(t, sampleWithoutSecrets, DeepCopyWithoutSecrets(sample))
}

func TestComplexStructureWithUnexported(t *testing.T) {
	sample := testSample()
	require.Equal(t, sample, DeepCopy(sample, OptionWithUnexported(true)))

	sampleWithoutSecrets := testSampleWithoutSecrets()
	require.Equal(t, sampleWithoutSecrets, DeepCopyWithoutSecrets(sample, OptionWithUnexported(true)))
}

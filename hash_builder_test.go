package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type s0T struct {
	a int
}

type s1T struct {
	a int
}

func TestCalcCryptoHash(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		s0 := s0T{a: 1}
		s0Dup := s0T{a: 1}
		s1 := s1T{a: 1}

		require.NotEqual(t, must(CalcCryptoHash(s0)), must(CalcCryptoHash(s1)))
		require.Equal(t, must(CalcCryptoHash(s0)), must(CalcCryptoHash(s0Dup)))
	})

	t.Run("map", func(t *testing.T) {
		m0 := map[string]int{
			"a": 0,
			"b": 1,
			"c": 2,
			"d": 3,
			"e": 4,
		}
		m0Dup := map[string]int{
			"a": 0,
			"b": 1,
			"c": 2,
			"d": 3,
			"e": 4,
		}
		m1 := map[string]int{
			"a": 0,
			"b": 1,
			"c": 2,
			"d": 3,
		}
		require.NotEqual(t, must(CalcCryptoHash(m0)), must(CalcCryptoHash(m1)))
		require.Equal(t, must(CalcCryptoHash(m0)), must(CalcCryptoHash(m0Dup)))
	})
}

func must[T any](in T, err error) T {
	if err != nil {
		panic(err)
	}
	return in
}

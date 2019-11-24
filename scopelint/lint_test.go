package scopelint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLintFiles(t *testing.T) {
	// This is not more than stopgap for issues.

	t.Run("#2", func(t *testing.T) {
		files := map[string][]byte{
			"mypkg/mypkg.go":      []byte("package mypkg\n"),
			"mypkg/mypkg_test.go": []byte("package mypkg_test"),
		}
		l := new(Linter)

		promblems, err := l.LintFiles(files)
		assert.NoError(t, err, "DO NOT make error for valid test package")
		assert.Empty(t, promblems)
	})
}

func TestLint(t *testing.T) {
	// This is not more than stopgap for issues.

	t.Run("#5: true positive", func(t *testing.T) {
		l := new(Linter)
		problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main

func factory() (ret func() *int) {
	for _, i := range make([]int, 1) {
		ret = func() *int { return &i }
	}
	return
}`))
		require.NoError(t, err)
		if assert.Len(t, problems, 2) {
			assert.Equal(t, "Using a reference for the variable on range scope \"i\"", problems[0].Text)
			assert.Equal(t, "Using the variable on range scope \"i\" in function literal", problems[1].Text)
		}
	})

	t.Run("#5: false positive", func(t *testing.T) {
		l := new(Linter)
		problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main

func returning() *int {
	for _, i := range make([]int, 1) {
		return &i
	}
	return nil
}`))
		require.NoError(t, err)
		assert.Empty(t, problems)
	})

	t.Run("issue #4", func(t *testing.T) {

		t.Run("positive", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main
import "testing"

func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			if "result" != tc.expected {
				t.Fatal("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 1) {
				assert.Equal(t, "Using the variable on range scope \"tc\" in function literal", problems[0].Text)
			}
		})

		t.Run("ignore line", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main

import "testing"

func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			t.Log(tc.expected) //scopelint:ignore // "result" != tc.expected
			if "result" != tc.expected {
				t.Fatal("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 2) {
				assert.True(t, problems[0].Ignored, "%#v", problems[0])
				assert.False(t, problems[1].Ignored, "%#v", problems[1])
			}
		})

		t.Run("ignore block", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main
import "testing"

func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			//scopelint:ignore
			if "result" != tc.expected {
				t.Fatal("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 1) {
				assert.True(t, problems[0].Ignored, "%#v", problems[0])
			}
		})

		t.Run("ignore block with other comments", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main
import "testing"

func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			//scopelint:ignore
			// compare expected and result
			if "result" != tc.expected {
				t.Fatal("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 1) {
				assert.True(t, problems[0].Ignored, "%#v", problems[0])
			}
		})

		t.Run("ignore ancestor block", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main
import "testing"

//scopelint:ignore
func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			if "result" != tc.expected {
				t.Fatal("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 1) {
				assert.True(t, problems[0].Ignored, "%#v", problems[0])
			}
		})
		t.Run("ignore file", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`//scopelint:ignore
package main

import "testing"

func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			if "result" != tc.expected {
				t.Fatal("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 1) {
				assert.True(t, problems[0].Ignored, "%#v", problems[0])
			}
		})

		t.Run("positive in next one of ignored line", func(t *testing.T) {
			l := new(Linter)
			problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main
import "testing"

func TestSomething(t *testing.T) {
	for _, tc := range []struct {
		expected    string
	}{} {
		t.Run("sub", func(t *testing.T) { // :memo: t.Run runs sub func immediately
			t.Log(tc.expected) //scopelint:ignore
			if "result" != tc.expected {
				t.Fatalf("failed")
			}
		})
	}
}`))

			require.NoError(t, err)
			if assert.Len(t, problems, 2) {
				assert.True(t, problems[0].Ignored, "%#v", problems[0])  // t.Log ~
				assert.False(t, problems[1].Ignored, "%#v", problems[1]) // if ~
			}
		})
	})
}

func TestLint_RangeLoop_ReferenceForStructField(t *testing.T) {
	t.Run("#1: reference for a struct field", func(t *testing.T) {
		l := new(Linter)
		problems, err := l.Lint("mypkg/mypkg.go", []byte(`package main

func factory() (out []*int) {
	for _, v := range make([]struct{ Field int }, 1) {
		out = append(out, &v.Field)
	}
	return out
}`))
		require.NoError(t, err)
		if assert.Len(t, problems, 1) {
			assert.Equal(t, "Using a reference for the variable on range scope \"v\"", problems[0].Text)
		}
	})
}

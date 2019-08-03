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
}

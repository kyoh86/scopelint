package scopelint

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

package scopelint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOptionComment(t *testing.T) {
	t.Run("empty comment", func(t *testing.T) {
		assert.Nil(t, parseOptionComment(""))
	})

	t.Run("normal comment", func(t *testing.T) {
		assert.Nil(t, parseOptionComment("comment"))
	})

	t.Run("only prefix", func(t *testing.T) {
		assert.Empty(t, parseOptionComment("scopelint:"))
	})

	t.Run("single option", func(t *testing.T) {
		assert.EqualValues(t, []string{"single"}, parseOptionComment("scopelint:single"))
	})

	t.Run("multiple option", func(t *testing.T) {
		assert.EqualValues(t, []string{"one", "two"}, parseOptionComment("scopelint:one,two"))
	})

	t.Run("multiple comment", func(t *testing.T) {
		assert.EqualValues(t, []string{"one", "two"}, parseOptionComment("scopelint:one//scopelint:two"))
	})

	t.Run("ignore spaces", func(t *testing.T) {
		assert.EqualValues(t, []string{"one", "two"}, parseOptionComment(" scopelint: 	 one 	 , two "))
	})
}

func TestHasOptionComment(t *testing.T) {
	assert.True(t, hasOptionComment("scopelint:one,two,three", "one"), "first one")
	assert.True(t, hasOptionComment("scopelint:one,two,three", "two"), "second one")
	assert.True(t, hasOptionComment("scopelint:one,two,three", "three"), "third one")
	assert.False(t, hasOptionComment("scopelint:one,two,three", "four"), "none")
}

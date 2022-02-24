package ans

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestANS_Send(t *testing.T) {
	t.Run("Straight forward test", func(t *testing.T) {
		err := Send("")
		assert.NoError(t, err)

	})
}

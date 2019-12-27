package gmailutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubjSplit(t *testing.T) {
	subjs := []string{
		`Новые статьи, связанные с работами автора Mohamed ...`,
		`"Learning to represent programs with graphs" - new citations`,
		`"machine learning on code" – de nouveaux résultats sont disponibles`,
	}

	for _, s := range subjs {
		srcType := NormalizeAndSplit(s)
		assert.Equal(t, 2, len(srcType), "%q parsing failed, restult: %+v", s, srcType)
	}
}

package gmailutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubjSplit(t *testing.T) {
	fixtures := []struct {
		subj       string
		src, typee string
	}{
		{
			`Новые статьи, связанные с работами автора Mohamed ...`,
			"Mohamed ...", related.En,
		},
		{
			`"Learning to represent programs with graphs" - new citations`,
			`"Learning to represent programs with graphs"`, citations.En,
		},
		{
			`"machine learning on code" – de nouveaux résultats sont disponibles`,
			`"machine learning on code"`, "de nouveaux résultats sont disponibles", // TODO(bzz): normalize FR so it's search.En here
		},
		{
			`Новые статьи пользователя Diomidis Spinellis`,
			"Diomidis Spinellis", articles.En,
		},
		{
			`Новые результаты по запросу "deep learning source code"`,
			`"deep learning source code"`, search.En,
		},
	}

	for _, f := range fixtures {
		srcType := NormalizeAndSplit(f.subj)
		assert.Equal(t, 2, len(srcType), "%q parsing failed, restult: %+v", f.subj, srcType)

		assert.Equal(t, f.src, srcType[0])
		assert.Equal(t, f.typee, srcType[1])
	}
}

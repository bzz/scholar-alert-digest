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
			`"Learning to represent programs with graphs"`, citationsOld.En,
		},
		{
			`3 new citations to articles by Diomidis Spinellis`,
			"Diomidis Spinellis", citations.En,
		},
		{
			`1 new citation to articles by Diomidis Spinellis`,
			"Diomidis Spinellis", citations.En,
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
		{
			`Diomidis Spinellis さんの論文からの引用: 123 件`,
			"Diomidis Spinellis", citations.En,
		},
		{
			`自分の論文からの引用: 1 件`,
			"自分", citations.En,
		},
		{
			`Diomidis Spinellis - 関連する新しい研究`,
			"Diomidis Spinellis", related.En,
		},
		{
			`Diomidis Spinellis - 新しい論文`,
			"Diomidis Spinellis", articles.En,
		},
		{
			`Diomidis Spinellis - 新しい結果`,
			"Diomidis Spinellis", search.En,
		},
		// {
		// 	`Рекомендуемые статьи`, "", recomended.En
		// }
	}

	for _, f := range fixtures {
		srcType := NormalizeAndSplit(f.subj)
		assert.Equal(t, 2, len(srcType), "%q parsing failed, restult: %+v", f.subj, srcType)

		assert.Equal(t, f.src, srcType[0])
		assert.Equal(t, f.typee, srcType[1])
	}
}

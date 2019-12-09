package papers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// UnitTests for paper extraction.

func TestScholarURLExtraction(t *testing.T) {
	var fixtures = []struct {
		scholarURL string
		URL        string
		toHaveErr  bool
	}{
		{ // error
			"", "", true,
		},
		{
			"http://scholar.google.com/scholar_url?url=https://arxiv.org/pdf/1911.12863&hl=en&sa=X&d=206864271411405978&scisig=AAGBfm07fPzie7SdYtYu_zrwxV7xx4o74g&nossl=1&oi=scholaralrt&hist=KBiQzPUAAAAJ:14254687125141938744:AAGBfm10na1baTgbjiNc57Wm9bK7bSlS3g",
			"https://arxiv.org/pdf/1911.12863", false,
		},
		{ // non .com
			"http://scholar.google.ru/scholar_url?url=https://www.jstage.jst.go.jp/article/transinf/E102.D/12/E102.D_2019MPP0005/_article/-char/ja/&hl=en",
			"https://www.jstage.jst.go.jp/article/transinf/E102.D/12/E102.D_2019MPP0005/_article/-char/ja/", false,
		},
		{
			"https://scholar.google.au/scholar_url?url=http://www.test.com&hl=1",
			"http://www.test.com", false,
		},
		{ // single query (no &)
			"http://scholar.google.au/scholar_url?url=http://www.test.com",
			"http://www.test.com", false,
		},
		{ // non-latin TLD
			"https://scholar.google.рф/scholar_url?url=http://www.test.com&hl=1",
			"http://www.test.com", false,
		},
	}

	for _, expected := range fixtures {
		actualURL, err := extractPaperURL(expected.scholarURL)
		if expected.toHaveErr {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, expected.URL, actualURL)
	}
}

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit test for the CLI utilities.

func TestScholarURLExtraction(t *testing.T) {
	var fixtures = []struct {
		scholarURL string
		URL        string
	}{
		{
			"http://scholar.google.com/scholar_url?url=https://arxiv.org/pdf/1911.12863&hl=en&sa=X&d=206864271411405978&scisig=AAGBfm07fPzie7SdYtYu_zrwxV7xx4o74g&nossl=1&oi=scholaralrt&hist=KBiQzPUAAAAJ:14254687125141938744:AAGBfm10na1baTgbjiNc57Wm9bK7bSlS3g",
			"https://arxiv.org/pdf/1911.12863",
		},
	}

	for _, expected := range fixtures {
		actualURL, err := extractPaperURL(expected.scholarURL)
		assert.Nil(t, err)
		assert.Equal(t, expected.URL, actualURL)
	}
}

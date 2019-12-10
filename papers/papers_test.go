package papers

import (
	"fmt"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UnitTests for paper extraction.

func TestScholarURLExtraction(t *testing.T) {
	var testCases = []struct {
		name       string
		scholarURL string
		URL        string
		hasErr     bool
	}{
		{
			"error", "", "", true,
		},
		{
			"regular .com",
			"http://scholar.google.com/scholar_url?url=https://arxiv.org/pdf/1911.12863&hl=en&sa=X&d=206864271411405978&scisig=AAGBfm07fPzie7SdYtYu_zrwxV7xx4o74g&nossl=1&oi=scholaralrt&hist=KBiQzPUAAAAJ:14254687125141938744:AAGBfm10na1baTgbjiNc57Wm9bK7bSlS3g",
			"https://arxiv.org/pdf/1911.12863", false,
		},
		{
			"non .com",
			"http://scholar.google.ru/scholar_url?url=https://www.jstage.jst.go.jp/article/transinf/E102.D/12/E102.D_2019MPP0005/_article/-char/ja/&hl=en",
			"https://www.jstage.jst.go.jp/article/transinf/E102.D/12/E102.D_2019MPP0005/_article/-char/ja/", false,
		},
		{
			"anothe TLD, short URL",
			"https://scholar.google.au/scholar_url?url=http://www.test.com&hl=1",
			"http://www.test.com", false,
		},
		{
			"single query (no &)",
			"http://scholar.google.au/scholar_url?url=http://www.test.com",
			"http://www.test.com", false,
		},
		{
			"non-latin TLD",
			"https://scholar.google.рф/scholar_url?url=http://www.test.com&hl=1",
			"http://www.test.com", false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualURL, err := extractPaperURL(tc.scholarURL)
			if tc.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.URL, actualURL)

		})
	}
}

var lineSplitCases = []struct {
	text         string
	n, lookahead int
	first, rest  string
}{
	{"й", 2, 2, "й", ""},
	{"abcd", 2, 2, "abcd", ""},
	{"abcdef", 2, 2, "abcd", "ef"},
	{"ab cdef", 2, 2, "ab", "cdef"},
	{
		"Многие методы преобразования программ (включая суперкомпиляцию и насыщение равенствами) можно сформулировать в виде набора правил переписывания графов или термов, применяемых в некотором порядке …",
		80, 10,
		"Многие методы преобразования программ (включая суперкомпиляцию и насыщение равенствами)",
		"можно сформулировать в виде набора правил переписывания графов или термов, применяемых в некотором порядке …",
	},
}

func TestAbstractFirstLine(t *testing.T) {
	for i, f := range lineSplitCases {
		first, rest := separateFirstLine(f.text, f.n, f.lookahead)
		require.Equal(t, f.first, first, "case %d", i)
		require.Equal(t, f.rest, rest, "case %d", i)
	}
}

func BenchmarkAbstractFirstLine(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, f := range lineSplitCases {
			separateFirstLine(f.text, f.n, f.lookahead)
		}
	}
}

// separateFirstLineSimple is presumably more readable version of first line separation.
//  BenchmarkAbstractFirstLine-4          	  2018 ns/op	       0 B/op	       0 allocs/op
//  BenchmarkAbstractFirstLineSimpler-4   	  6008 ns/op	    1176 B/op	      14 allocs/op
func separateFirstLineSimple(text string, N, lookahead int) (string, string) {
	text = strings.ReplaceAll(text, "\n", "")

	total := 0
	var first []string
	for _, w := range strings.Fields(text) {
		total += utf8.RuneCountInString(w)
		if total > N {
			break
		}
		first = append(first, w)
	}

	f := strings.Join(first, " ")
	if f == "" {
		r := []rune(text)
		if len(r) > N+lookahead {
			f = string(r[:N+lookahead])
		} else {
			f = text
		}
	}

	r := strings.TrimPrefix(text, f)
	return f, strings.TrimLeftFunc(r, unicode.IsSpace)
}

func TestAbstractFirstLineSimpler(t *testing.T) {
	for i, f := range lineSplitCases {
		t.Run(fmt.Sprintf("case %d", i), func(*testing.T) {
			first, rest := separateFirstLineSimple(f.text, f.n, f.lookahead)
			require.Equal(t, f.first, first)
			require.Equal(t, f.rest, rest)
		})
	}
}

func BenchmarkAbstractFirstLineSimpler(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, f := range lineSplitCases {
			separateFirstLineSimple(f.text, f.n, f.lookahead)
		}
	}
}

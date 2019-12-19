// Package papers indes a data model and utilities for paper details extraction.
package papers

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/antchfx/htmlquery"
	"google.golang.org/api/gmail/v1"

	"github.com/bzz/scholar-alert-digest/gmailutils"
)

var scholarURLPrefix = regexp.MustCompile(`http(s)?://scholar\.google\.\p{L}+/scholar_url\?url=`)

// Paper is a map key, thus aggregation take into account all it's fields.
type Paper struct {
	// TODO(bzz):
	// Freq       int
	Title    string
	URL      string
	Author   string
	Abstract Abstract
	// Refs       []string
}

// Abstract represents a view of the parsed abstract.
type Abstract struct {
	FirstLine, Rest string
}

// AggPapers represents an aggregated collection of Papers.
// TODO(bzz): think about aggregation only by the title, as suggested in
// https://github.com/bzz/scholar-alert-digest/issues/12#issuecomment-562820924
// e.g AggPapers [string]*Paper + Paper.Freq
type AggPapers map[Paper]int

// Stats is a number of counters \w stats on paper extraction from gmail messages.
type Stats struct {
	Msgs, Titles, Errs int
}

// Helpers for a Map, sorted by keys.
// TODO(bzz): move to map.go after `go run main.go` is replaced by ./cmd/report
type sortedMap struct {
	m map[Paper]int
	s []Paper
}

func (sm *sortedMap) Len() int           { return len(sm.m) }
func (sm *sortedMap) Less(i, j int) bool { return sm.m[sm.s[i]] > sm.m[sm.s[j]] }
func (sm *sortedMap) Swap(i, j int)      { sm.s[i], sm.s[j] = sm.s[j], sm.s[i] }

// SortedKeys sort the given map by key.
// TODO(bzz): use a stable sort
func SortedKeys(m map[Paper]int) []Paper {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]Paper, len(m))
	i := 0
	for key := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

// ExtractPapersFromMsgs parses the messages payloads and creates Papers.
// TODO(bzz): rename to extract+aggregate
func ExtractPapersFromMsgs(msgs []*gmail.Message, inclAuthors bool) (*Stats, AggPapers) {
	st := &Stats{Msgs: len(msgs)}
	uniqTitles := AggPapers{}

	for _, m := range msgs {
		papers, err := extractPapersFromMsg(m, inclAuthors)
		if err != nil {
			st.Errs++
			continue
		}

		// aggregate
		// TODO(bzz): reaplace by title aggregation (AggPapers [string]*Paper)
		st.Titles += len(papers)
		for _, paper := range papers {
			uniqTitles[paper]++
		}
	}

	return st, uniqTitles
}

func extractPapersFromMsg(m *gmail.Message, inclAuthors bool) ([]Paper, error) {
	subj := gmailutils.Subject(m.Payload)

	body, err := gmailutils.MessageTextBody(m)
	if err != nil {
		e := fmt.Errorf("failed to get message text for ID %s - %s", m.Id, err)
		return nil, e
	}

	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		e := fmt.Errorf("failed to parse HTML body of %q", subj)
		return nil, e
	}

	// paper titles, from a single email
	xpTitle := "//h3/a"
	titles, err := htmlquery.QueryAll(doc, xpTitle)
	if err != nil {
		return nil, fmt.Errorf("title: not valid XPath expression %q", xpTitle)
	}

	// paper urls, from a single email
	xpURL := "//h3/a/@href"
	urls, err := htmlquery.QueryAll(doc, xpURL)
	if err != nil {
		return nil, fmt.Errorf("url: not valid XPath expression %q", xpURL)
	}

	if len(titles) != len(urls) {
		e := fmt.Errorf("titles %d != %d urls in %q", len(titles), len(urls), subj)
		return nil, e
	}

	// paper authors & year
	xpAuth := "//h3/following-sibling::div[1]"
	auths, err := htmlquery.QueryAll(doc, xpAuth)
	if err != nil {
		return nil, fmt.Errorf("authors: not valid XPath expression %q", xpAuth)
	}

	// paper abstract
	xpAbs := "//h3/following-sibling::div[2]"
	abss, err := htmlquery.QueryAll(doc, xpAbs)
	if err != nil {
		return nil, fmt.Errorf("abstract: not valid XPath expression %q", xpAbs)
	}

	var papers []Paper
	var author string
	for i, aTitle := range titles {
		title := strings.TrimSpace(htmlquery.InnerText(aTitle))
		abstract := strings.TrimSpace(htmlquery.InnerText(abss[i]))
		if inclAuthors {
			author = extractPaperAuthor(htmlquery.InnerText(auths[i]))
		}

		url, err := extractPaperURL(htmlquery.InnerText(urls[i]))
		if err != nil {
			log.Printf("Skipping paper %q in %q: %s", title, subj, err)
			continue
		}

		N, lookahead := 80, 10 // max number of runes to process
		first, rest := separateFirstLine(abstract, N, lookahead)
		papers = append(papers, Paper{title, url, author, Abstract{first, rest}})
	}
	return papers, nil
}

func extractPaperAuthor(publication string) string {
	auth := publication
	for i, r := range publication {
		if unicode.In(r, unicode.Dash) {
			auth = strings.TrimRightFunc(publication[:i], unicode.IsSpace)
			break
		}
	}
	return strings.Title(strings.ToLower(auth))
}

// extractPaperURL returns an actual paper URL from the given scholar link.
// Does not validate URL format but extracts it ad-hoc by trimming sufix/prefix.
func extractPaperURL(scholarURL string) (string, error) {
	// drop scholarURLPrefix
	prefixLoc := scholarURLPrefix.FindStringIndex(scholarURL)
	if prefixLoc == nil {
		return "", fmt.Errorf("url %q does not have prefix %q", scholarURL, scholarURLPrefix.String())
	}
	longURL := scholarURL[prefixLoc[1]:]

	// drop sufix (after &), if any
	if sufix := strings.Index(longURL, "&"); sufix >= 0 {
		longURL = longURL[:sufix]
	}

	return url.QueryUnescape(longURL)
}

// separateFirstLine returns text, split into two parts: first short line and the rest.
// N+lookehead is max length of the first. Split is done unicode whitespace,
// if any around N +/-lookahead runes, or at Nth rune.
func separateFirstLine(text string, N, lookahead int) (string, string) {
	text = strings.ReplaceAll(text, "\n", "")
	if len(text) < N { // bytes, not code-points
		return text, ""
	}

	pos := 0
	n, nPos, lastSpace, lastSpacePos := 0, 0, 0, 0
	for len(text[pos:]) > 0 {
		if n >= N+lookahead {
			break
		}
		char, width := utf8.DecodeRuneInString(text[pos:])
		pos += width
		if unicode.IsSpace(char) {
			lastSpacePos = pos
			lastSpace = n
		}
		n, nPos = n+1, pos
	}

	cut := nPos
	if abs(N-lastSpace) < lookahead { // whitespace in lookahead neighborhood of Nth rune
		cut = lastSpacePos
	}
	return strings.TrimRightFunc(text[:cut], unicode.IsSpace), text[cut:]
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

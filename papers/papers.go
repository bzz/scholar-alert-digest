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

	"github.com/antchfx/htmlquery"
	"google.golang.org/api/gmail/v1"

	"github.com/bzz/scholar-alert-digest/gmailutils"
)

var scholarURLPrefix = regexp.MustCompile(`http(s)?://scholar\.google\.\p{L}+/scholar_url\?url=`)

// Paper is a map key, thus aggregation take into account all it's fields.
//
// TODO(bzz): think about aggregation only by the title, as suggested in
// https://github.com/bzz/scholar-alert-digest/issues/12#issuecomment-562820924
type Paper struct {
	Title, URL string
	Abstract   Abstract
}

// Abstract represents a view of parsed abstract.
type Abstract struct {
	Full, FirstLine, RestLines string
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
func ExtractPapersFromMsgs(messages []*gmail.Message) (int, int, map[Paper]int) {
	errCount := 0
	titlesCount := 0
	uniqTitles := map[Paper]int{}

	for _, m := range messages {
		papers, err := extractPapersFromMsg(m)
		if err != nil {
			errCount++
			continue
		}

		titlesCount += len(papers)
		for _, paper := range papers { // map title to uniqTitles
			uniqTitles[paper]++
		}
	}

	return errCount, titlesCount, uniqTitles
}

func extractPapersFromMsg(m *gmail.Message) ([]Paper, error) {
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

	// paper abstract
	xpAbs := "//h3/following-sibling::div[2]"
	abss, err := htmlquery.QueryAll(doc, xpAbs)
	if err != nil {
		return nil, fmt.Errorf("abstract: not valid XPath expression %q", xpAbs)
	}

	var papers []Paper
	for i, aTitle := range titles {
		title := strings.TrimSpace(htmlquery.InnerText(aTitle))
		abs := strings.TrimSpace(htmlquery.InnerText(abss[i]))

		url, err := extractPaperURL(htmlquery.InnerText(urls[i]))
		if err != nil {
			log.Printf("Skipping paper %q in %q: %s", title, subj, err)
			continue
		}

		papers = append(papers, Paper{
			title, url, Abstract{
				abs, separateFirstLine(abs)[0], separateFirstLine(abs)[1],
			},
		})
	}
	return papers, nil
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

func separateFirstLine(text string) []string {
	text = strings.ReplaceAll(text, "\n", "")
	n := 80 // TODO(bzz): utf8 whitespace-aware splitting alg capped by max N runes
	if len(text) < n {
		return []string{text, ""}
	}
	return []string{text[:n], text[n:]}
}

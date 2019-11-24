/**
 * Copyright 2019 Alexander Bezzubov.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// CLI tool for aggregating unread messages in Gmail from Google Scholar Alert.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bzz/scholar-alert-digest/gmailutils"

	"github.com/antchfx/htmlquery"
	"google.golang.org/api/gmail/v1"
)

const (
	labelName  = "[-oss-]-_ml-in-se" // "[ OSS ]/_ML-in-SE" in the Web UI
	scholarURL = "http://scholar.google.com/scholar_url?url="

	usageMessage = `usage: go run [-labels] [-l <your-gmail-label>]

Polls Gmail API for unread Google Scholar alert messaged under a given label,
aggregates by paper title and prints a list of paper URLs in Markdown format.

The -labels flag will only list all available labels for the current account.
`
)

var (
	user = "me"

	gmailLabel = flag.String("l", labelName, "name of the Gmail label")
	listLabels = flag.Bool("labels", false, "list all Gmail labels")
)

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(0)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	client := gmailutils.NewClient()
	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create a Gmail client: %v", err)
	}

	if *listLabels {
		gmailutils.PrintAllLabels(srv, user)
		os.Exit(0)
	}

	// TODO(bzz): fetchGmailMsgsAsync returning chan *gmail.Message
	start := time.Now()
	var messages []*gmail.Message = fetchGmailMsgs(srv, user, *gmailLabel)
	log.Printf("%d unread messages found (took %.0f sec)", len(messages), time.Since(start).Seconds())

	errCount, titlesCount, uniqTitles := extractPapersFromMsgs(messages)

	generateMarkdownReport(len(messages), titlesCount, uniqTitles)

	// TODO(bzz): add state
	//  update report from FS \w checkbox state, instead of always generating a new one
	//  mark emails as "read", when all the links are checked off

	if errCount != 0 {
		log.Printf("Errors: %d\n", errCount)
	}
}

func generateMarkdownReport(messagesCount, titlesCount int, uniqTitles map[paper]int) {
	// TODO(bzz): convert to text/template, as soon as it becomes more complex
	fmt.Printf("Date: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Unread emails: %d\n", messagesCount)
	fmt.Printf("Paper titles: %d\n", titlesCount)
	fmt.Printf("Uniq paper titles: %d\n\n", len(uniqTitles))

	// TODO(bzz): use a stable sort
	for _, paper := range sortedKeys(uniqTitles) {
		fmt.Printf(" - [ ] [%s](%s) (%d)\n", paper.title, paper.url, uniqTitles[paper])
	}
}

// fetchGmailMsgs fetches all unread messages under a certain lable from Gmail.
func fetchGmailMsgs(srv *gmail.Service, user, label string) []*gmail.Message {
	if envLabel, ok := os.LookupEnv("SAD_LABEL"); ok {
		gmailLabel = &envLabel
	}

	return gmailutils.UnreadMessagesInLabel(srv, user, label)
}

func extractPapersFromMsgs(messages []*gmail.Message) (int, int, map[paper]int) {
	errCount := 0
	titlesCount := 0
	uniqTitles := map[paper]int{}

	for _, m := range messages {
		papers, err := extractAllPapers(m)
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

func extractAllPapers(m *gmail.Message) ([]paper, error) {
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

	var papers []paper
	for i, aTitle := range titles {
		title := strings.TrimSpace(htmlquery.InnerText(aTitle))

		longURL := strings.TrimPrefix(htmlquery.InnerText(urls[i]), scholarURL)
		url, err := url.QueryUnescape(longURL[:strings.Index(longURL, "&")])
		if err != nil {
			log.Printf("Skipping paper %q in %q: %s", title, subj, err)
			continue
		}

		papers = append(papers, paper{title, url})
	}
	return papers, nil
}

// Helpers for a Map, sorted by keys.
// TODO(bzz): move to map.go after `go run main.go` is replaced by ./cmd/report
type sortedMap struct {
	m map[paper]int
	s []paper
}

func (sm *sortedMap) Len() int           { return len(sm.m) }
func (sm *sortedMap) Less(i, j int) bool { return sm.m[sm.s[i]] > sm.m[sm.s[j]] }
func (sm *sortedMap) Swap(i, j int)      { sm.s[i], sm.s[j] = sm.s[j], sm.s[i] }

func sortedKeys(m map[paper]int) []paper {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]paper, len(m))
	i := 0
	for key := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

type paper struct {
	title, url string
}

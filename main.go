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
aggregates by paper titles and prints a list of paper URLs in Markdown.

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
		log.Printf("Listing all Gmail labels")
		lablesResp, err := srv.Users.Labels.List(user).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve all labels: %v", err)
		}

		log.Printf("%d labels found", len(lablesResp.Labels))
		for _, label := range lablesResp.Labels {
			fmt.Printf("%s\n", label.Name)
		}
		os.Exit(0)
	}

	var messages []*gmail.Message = gmailutils.UnreadMessagesInLabel(srv, user, *gmailLabel)

	log.Printf("%d unread messages found", len(messages))
	errCount := 0
	totalTitles := 0
	uniqTitlesCount := map[paper]int{}
	for _, m := range messages {
		subj := gmailutils.Subject(m.Payload)

		body, err := gmailutils.MessageTextBody(m)
		if err != nil {
			fmt.Printf("Failed to get message text for ID %s - %s\n", m.Id, err)
			errCount++
			continue
		}

		doc, err := htmlquery.Parse(bytes.NewReader(body))
		if err != nil {
			fmt.Printf("Failed to parse HTML body of %q\n", subj)
			errCount++
			continue
		}

		// paper titles, from a single email
		xpTitle := "//h3/a"
		titles, err := htmlquery.QueryAll(doc, xpTitle)
		if err != nil {
			fmt.Printf("Not valid XPath expression %q\n", xpTitle)
		}
		totalTitles += len(titles)

		// paper urls, from a single email
		xpURL := "//h3/a/@href"
		urls, err := htmlquery.QueryAll(doc, xpURL)
		if err != nil {
			fmt.Printf("Not valid XPath expression %q\n", xpURL)
		}

		if len(titles) != len(urls) {
			fmt.Printf("Titles %d != %d urls in %q. Skipping email\n", len(titles), len(urls), subj)
			errCount++
			continue
		}

		for i, aTitle := range titles { // titles -> uniqTitlesCount
			title := strings.TrimSpace(htmlquery.InnerText(aTitle))

			u := strings.TrimPrefix(htmlquery.InnerText(urls[i]), scholarURL)
			url, err := url.QueryUnescape(u[:strings.Index(u, "&")])
			if err != nil {
				fmt.Printf("Skipping paper %q: %s\n", title, err)
				continue
			}

			p := paper{title, url}
			uniqTitlesCount[p]++
		}
	}

	fmt.Printf("Date: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Unread emails: %d\n", len(messages))
	fmt.Printf("Paper titles: %d\n", totalTitles)
	fmt.Printf("Uniq paper titles: %d\n\n", len(uniqTitlesCount))

	// TODO(bzz):
	//  update existing report \w checkbox state, instead of always generating a new one
	//  mark emails as "read", when all the links are checked off

	// generate Markdown report
	for _, paper := range sortedKeys(uniqTitlesCount) {
		fmt.Printf(" - [ ] [%s](%s) (%d)\n", paper.title, paper.url, uniqTitlesCount[paper])
	}

	if errCount != 0 {
		fmt.Printf("Errors: %d\n", errCount)
	}
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

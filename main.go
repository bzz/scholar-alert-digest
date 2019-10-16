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

	"github.com/bzz/scholar-alert-gmail-digest/gmailutils"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/context"
	"google.golang.org/api/gmail/v1"
)

const (
	labelName  = "[-oss-]-_ml_in_se"
	scholarURL = "http://scholar.google.com/scholar_url?url="

	usageMessage = `usage: go run [-l <your-gmail-label>]

Polls unread Google Scholar messaged under a given label from GMail though the API
aggregates them by paper and outputs a Markdown list of paper URLs.
`
)

var (
	user  = "me"
	query = fmt.Sprintf("label:%s is:unread", labelName)

	gmailLabel = flag.String("l", labelName, "write cpu profile to this file")
)

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(2)
}

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

func main() {
	flag.Usage = usage
	flag.Parse()

	client := gmailutils.NewClient()
	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	log.Printf("Searching for all unread messages under Gmail label %q", *gmailLabel)
	page := 0
	var messages []*gmail.Message
	err = srv.Users.Messages.List(user).Q(query).Pages(context.TODO(), func(rm *gmail.ListMessagesResponse) error {
		for _, m := range rm.Messages {
			msg, err := srv.Users.Messages.Get(user, m.Id).Do()
			if err != nil {
				return err
			}

			messages = append(messages, msg)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to retrieve messages with query %q, page %d: %v", query, page, err)
	}

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

		// get paper titles
		xpTitle := "//h3/a"
		titles, err := htmlquery.QueryAll(doc, xpTitle)
		if err != nil {
			fmt.Printf("Not valid XPath expression %q\n", xpTitle)
		}
		totalTitles += len(titles)

		// get paper urls
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

		// fmt.Printf(" - %s (%d)\n", subj, len(paperTitles))
		for i, aTitle := range titles {
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

	fmt.Printf("Total paper titles: %d\n", totalTitles)
	fmt.Printf("Uniq paper titles: %d\n", len(uniqTitlesCount))

	// TODO(bzz): generate Markdown
	for _, paper := range sortedKeys(uniqTitlesCount) {
		fmt.Printf(" - [%s](%s) (%d)\n", paper.title, paper.url, uniqTitlesCount[paper])
	}

	if errCount != 0 {
		fmt.Printf("Errors: %d\n", errCount)
	}
}

/**
 * Copyright Google Inc.
 * Copyright 2019 Alexander Bezzubov
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
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bzz/scholar-alert-gmail-digest/gmailutils"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/context"
	"google.golang.org/api/gmail/v1"
)

var (
	user      = "me"
	labelName = "[-oss-]-_ml_in_se"
	query     = fmt.Sprintf("label:%s is:unread", labelName)
)

type sortedMap struct {
	m map[string]int
	s []string
}

func (sm *sortedMap) Len() int           { return len(sm.m) }
func (sm *sortedMap) Less(i, j int) bool { return sm.m[sm.s[i]] > sm.m[sm.s[j]] }
func (sm *sortedMap) Swap(i, j int)      { sm.s[i], sm.s[j] = sm.s[j], sm.s[i] }

func sortedKeys(m map[string]int) []string {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]string, len(m))
	i := 0
	for key := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

func main() {
	client := gmailutils.NewClient()
	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

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

	errCount := 0
	totalTitles := 0
	uniqTitlesCount := map[string]int{}
	fmt.Printf("Un-read messages: %d\n", len(messages))
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

		// text //h3/a
		xp := "//h3/a"
		paperTitles, err := htmlquery.QueryAll(doc, xp)
		if err != nil {
			fmt.Printf("Not valid XPath expression %q\n", xp)
		}
		// fmt.Printf(" - %s (%d)\n", subj, len(paperTitles))
		totalTitles += len(paperTitles)

		for _, titleA := range paperTitles {
			t := strings.TrimSpace(htmlquery.InnerText(titleA))
			uniqTitlesCount[t]++
		}
	}

	fmt.Printf("Total paper titles: %d\n", totalTitles)
	fmt.Printf("Uniq paper titles: %d\n", len(uniqTitlesCount))

	// TODO(bzz): generate Markdown
	// url: parse "//h3/a/@href" up to '&', drop refix "http://scholar.google.com/scholar_url?"
	for _, title := range sortedKeys(uniqTitlesCount) {
		fmt.Printf(" - %s (%d)\n", title, uniqTitlesCount[title])
	}

	if errCount != 0 {
		fmt.Printf("Errors: %d\n", errCount)
	}
}

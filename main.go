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
//
// It does so by:
//  - fetching messages under a certian Gmail label
//  - transforming and aggregateing them into map[paper]int
//  - rendering a text/template with it, in Markdown/HTML/JSON
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bzz/scholar-alert-digest/gmailutils"
	"github.com/bzz/scholar-alert-digest/papers"
	"github.com/bzz/scholar-alert-digest/templates"

	"google.golang.org/api/gmail/v1"
)

const (
	labelName = "[-oss-]-_ml-in-se" // "[ OSS ]/_ML-in-SE" in the Web UI

	usageMessage = `usage: go run [-labels | -subj] [-html | -json] [-compact] [-mark] [-read] [-authors] [-refs] [-l <your-gmail-label>] [-n]

Polls Gmail API for unread Google Scholar alert messaged under a given label,
aggregates by paper title and prints a list of paper URLs in Markdown format.

The -l flag sets the Gmail label to look for (overriden by 'SAD_LABEL' env variable).
The -n flag sets the number of concurent requests to Gmail API.
The -labels flag will only print all available labels for the current account.
The -subj flag will only include email subjects in the report. Usefull for " | uniq -c | sort -dr".
The -html flag will produce ouput report in HTML format.
The -json flag will produce output in JSONL format, one paper object per line.
The -compact flag will produce ouput report in compact format, usefull >100 papers.
The -mark flag will mark all the aggregated emails as read in Gmail.
The -read flag will include a new section in the report, aggregating all read emails.
The -authors flag will include paper authors in the report.
The -refs flag will add links to all email messages that mention each paper.
`
)

var (
	user = "me" // TODO(bzz): move to const in gmailutils

	gmailLabel = flag.String("l", labelName, "name of the Gmail label")
	listLabels = flag.Bool("labels", false, "list all Gmail labels")
	// TODO(bzz): a format flag \w validated md/html/json options would be better
	outputHTML = flag.Bool("html", false, "output report in HTML (instead of default Markdown)")
	outputJSON = flag.Bool("json", false, "output report data in JSON")
	compact    = flag.Bool("compact", false, "output report in compact format (>100 papers)")
	markRead   = flag.Bool("mark", false, "marks all aggregated emails as read")
	read       = flag.Bool("read", false, "include read emails to a separate section of the report")
	authors    = flag.Bool("authors", false, "include paper authors in the report")
	refs       = flag.Bool("refs", false, "include orignin references to Gmail messages in report")
	onlySubj   = flag.Bool("subj", false, "aggregate only email subjects")
	concurReq  = flag.Int("n", 10, "number of concurent Gmail API requests")
)

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	os.Exit(0)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	client := gmailutils.NewClient(*markRead)
	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create a Gmail client: %v", err)
	}

	if *listLabels {
		gmailutils.PrintAllLabels(srv, user)
		os.Exit(0)
	}

	// override lable name by env var
	if envLabel, ok := os.LookupEnv("SAD_LABEL"); ok {
		gmailLabel = &envLabel
	}

	if *onlySubj {
		log.Print("only extracting the subjects")
		query := fmt.Sprintf("label:%s is:unread", *gmailLabel)
		if *read {
			query = strings.TrimSuffix(query, " is:unread")
		}

		msgs, err := gmailutils.FetchConcurent(context.Background(), srv, user, query, *concurReq)
		if err != nil {
			log.Fatalf("Failed to fetch messages from Gmail: %v", err)
		}

		printSubjects(msgs)
		os.Exit(0)
	}

	// fetch messages, extract papers, aggregated by title
	// TODO(bzz): FetchAsync returning chan *gmail.Message?
	urMsgs, err := gmailutils.FetchConcurent(context.Background(), srv, user, fmt.Sprintf("label:%s is:unread", *gmailLabel), *concurReq)
	if err != nil {
		log.Fatalf("Failed to fetch messages from Gmail: %v", err)
	}
	unreadStats, unreadPapers := papers.ExtractAndAggPapersFromMsgs(urMsgs, *authors, *refs)

	readStats := &papers.Stats{}
	var readPapers papers.AggPapers
	if *read {
		rMsgs, err := gmailutils.FetchConcurent(context.Background(), srv, user, fmt.Sprintf("label:%s is:read", *gmailLabel), *concurReq)
		if err != nil {
			log.Fatal("Failed to fetch messages from Gmail")
		}
		readStats, readPapers = papers.ExtractAndAggPapersFromMsgs(rMsgs, *authors, *refs)
	}

	// render papers
	var r templates.Renderer
	template, style := templates.MdTemplText, ""
	if *compact {
		template, style = templates.CompactMdTemplText, templates.CompatStyle
	}

	log.Printf("rendering %d papers", len(unreadPapers)+len(readPapers))
	if *outputJSON {
		r = templates.NewJSONRenderer()
	} else if *outputHTML {
		r = templates.NewHTMLRenderer(template, style)
	} else {
		r = templates.NewMarkdownRenderer(template, templates.ReadMdTemplText)
	}
	r.Render(os.Stdout, unreadStats, unreadPapers, readPapers)

	if *markRead {
		// TODO(bzz): add a state
		//  use existing report from FS \w a checkbox state set by the user
		//  only mark email as "read" iff all the links are checked off
		gmailutils.ModifyMsgsDelLabel(srv, user, urMsgs, "UNREAD")
	}

	totalErrCnt := unreadStats.Errs + readStats.Errs
	if totalErrCnt != 0 {
		log.Printf("Errors: %d\n", totalErrCnt)
	}
}

func printSubjects(msgs []*gmail.Message) {
	var subjs []string
	for _, m := range msgs {
		subj := gmailutils.Subject(m.Payload)
		srcType, sep := splitOnDash(subj)
		if len(srcType) != 2 {
			log.Printf("subject %q can not be split by %q", subj, sep)
			continue
		}

		subjs = append(subjs, fmt.Sprintf("%-22s | %s", srcType[1], srcType[0]))
	}
	sort.Strings(subjs)
	for _, s := range subjs {
		fmt.Printf("%s\n", s)
	}
}

// splitOnDash returns str, split on unicode Dash and a separator.
func splitOnDash(str string) ([]string, string) {
	s := str
	dash := "-"
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		if unicode.In(r, unicode.Dash) {
			dash = string(r)
			break
		}
	}
	sep := fmt.Sprintf(" %s ", dash)
	return strings.Split(str, sep), sep
}

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
//  - rendering a text/template with it, in Markdown or HTML
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/bzz/scholar-alert-digest/gmailutils"

	"github.com/antchfx/htmlquery"
	"gitlab.com/golang-commonmark/markdown"
	"google.golang.org/api/gmail/v1"
)

const (
	labelName = "[-oss-]-_ml-in-se" // "[ OSS ]/_ML-in-SE" in the Web UI

	usageMessage = `usage: go run [-labels] [-html] [-mark] [-read] [-l <your-gmail-label>]

Polls Gmail API for unread Google Scholar alert messaged under a given label,
aggregates by paper title and prints a list of paper URLs in Markdown format.

The -l flag sets the Gmail label to look for (overriden by 'SAD_LABEL' env variable).
The -labels flag will only list all available labels for the current account.
The -html flag will produce ouput report in HTML format.
The -mark flag will mark all the aggregated emails as read in Gmail.
The -read flag will include a new section in the report, aggregating all read emails.
`

	newMdTemplText = `# Google Scholar Alert Digest

**Date**: {{.Date}}
**Unread emails**: {{.UnreadEmails}}
**Paper titles**: {{.TotalPapers}}
**Uniq paper titles**: {{.UniqPapers}}

## New papers
{{ range $paper := sortedKeys .Papers }}
 - [{{ .Title }}]({{ .URL }}) ({{index $.Papers .}})
   {{- if .Abstract.Full }}
   <details>
     <summary>{{.Abstract.FirstLine}}</summary>{{.Abstract.RestLines}}
   </details>
   {{ end }}
{{ end }}
`

	oldMdTemplText = `## Old papers

<details>
  <summary>Archive</summary>

{{ range $paper := sortedKeys . }}
  - [{{ .Title }}]({{ .URL }})
    {{- if .Abstract.Full }}
    <details>
      <summary>{{.Abstract.FirstLine}}</summary>{{.Abstract.RestLines}}
    </details>
    {{ end }}
{{ end }}
</details>
`

	htmlTemplText = `<!DOCTYPE html>
<html lang="en">
  <head><meta charset="UTF-8"></head>
  <body>%s</body>
</html>
`
)

var (
	user             = "me"
	scholarURLPrefix = regexp.MustCompile(`http(s)?://scholar\.google\.\p{L}+/scholar_url\?url=`)

	gmailLabel = flag.String("l", labelName, "name of the Gmail label")
	listLabels = flag.Bool("labels", false, "list all Gmail labels")
	// TODO(bzz): a format flag \w validated md/html options would be better
	ouputHTML = flag.Bool("html", false, "output report in HTML (instead of default Markdown)")
	markRead  = flag.Bool("mark", false, "marks all aggregated emails as read")
	read      = flag.Bool("read", false, "include read emails to a separate section of the report")
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

	// TODO(bzz): fetchGmailAsync returning chan *gmail.Message
	var urMsgs []*gmail.Message = fetchGmail(srv, user, fmt.Sprintf("label:%s is:unread", *gmailLabel))
	errCnt, urTitlesCnt, urTitles := extractPapersFromMsgs(urMsgs)

	var rTitles map[paper]int
	if *read {
		rMsgs := fetchGmail(srv, user, fmt.Sprintf("label:%s is:read", *gmailLabel))
		_, _, rTitles = extractPapersFromMsgs(rMsgs)
	}

	if *ouputHTML {
		generateAndPrintHTML(len(urMsgs), urTitlesCnt, urTitles, rTitles)
	} else {
		generateAndPrintMarkdown(len(urMsgs), urTitlesCnt, urTitles, rTitles)
	}

	if *markRead {
		// TODO(bzz): add a state
		//  use existing report from FS \w a checkbox state set by the user
		//  only mark email as "read" iff all the links are checked off
		markGmailMsgsUnread(srv, user, urMsgs)
	}

	if errCnt != 0 {
		log.Printf("Errors: %d\n", errCnt)
	}
}

// fetchGmail fetches all messages matching a given query from the Gmail.
func fetchGmail(srv *gmail.Service, user, query string) []*gmail.Message {
	start := time.Now()
	msgs := gmailutils.QueryMessages(srv, user, query)
	log.Printf("%d messages found under %q (took %.0f sec)", len(msgs), query, time.Since(start).Seconds())
	return msgs
}

func extractPapersFromMsgs(messages []*gmail.Message) (int, int, map[paper]int) {
	errCount := 0
	titlesCount := 0
	uniqTitles := map[paper]int{}

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

func extractPapersFromMsg(m *gmail.Message) ([]paper, error) {
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

	var papers []paper
	for i, aTitle := range titles {
		title := strings.TrimSpace(htmlquery.InnerText(aTitle))
		abs := strings.TrimSpace(htmlquery.InnerText(abss[i]))

		url, err := extractPaperURL(htmlquery.InnerText(urls[i]))
		if err != nil {
			log.Printf("Skipping paper %q in %q: %s", title, subj, err)
			continue
		}

		papers = append(papers, paper{
			title, url, abstract{
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

func generateAndPrintHTML(msgsCnt, titlesCnt int, unread, read map[paper]int) {
	var mdBuf bytes.Buffer
	generateMarkdown(&mdBuf, msgsCnt, titlesCnt, unread, read)

	md := markdown.New(markdown.XHTMLOutput(true), markdown.HTML(true))
	fmt.Printf(htmlTemplText, md.RenderToString([]byte(mdBuf.String())))
}

func generateAndPrintMarkdown(msgsCnt, titlesCnt int, unread, read map[paper]int) {
	generateMarkdown(os.Stdout, msgsCnt, titlesCnt, unread, read)
}

func generateMarkdown(out io.Writer, msgsCnt, titlesCnt int, unread, read map[paper]int) {
	newMdReport(out, msgsCnt, titlesCnt, unread)
	if read != nil {
		oldMdReport(out, read)
	}
}

// renderes newMdTemplText for new, unread papers.
func newMdReport(out io.Writer, msgsCnt, titlesCnt int, papers map[paper]int) {
	tmplText := newMdTemplText
	tmpl := template.Must(template.New("unread-papers").Funcs(template.FuncMap{
		"sortedKeys": sortedKeys,
	}).Parse(tmplText))
	err := tmpl.Execute(out, struct {
		Date         string
		UnreadEmails int
		TotalPapers  int
		UniqPapers   int
		Papers       map[paper]int
	}{
		time.Now().Format(time.RFC3339),
		msgsCnt,
		titlesCnt,
		len(papers),
		papers,
	})
	if err != nil {
		log.Fatalf("template %q execution failed: %s", tmplText, err)
	}
}

// renderes oldMdTemplText for old, read papers
func oldMdReport(out io.Writer, papers map[paper]int) {
	tmplText := oldMdTemplText
	tmpl := template.Must(template.New("read-papers").Funcs(template.FuncMap{
		"sortedKeys": sortedKeys,
	}).Parse(tmplText))
	err := tmpl.Execute(out, papers)
	if err != nil {
		log.Fatalf("template %q execution failed: %s", tmplText, err)
	}
}

func markGmailMsgsUnread(srv *gmail.Service, user string, messages []*gmail.Message) {
	const label = "UNREAD"
	var msgIds []string
	for _, msg := range messages {
		msgIds = append(msgIds, msg.Id)
	}

	err := srv.Users.Messages.BatchModify(user, &gmail.BatchModifyMessagesRequest{
		Ids:            msgIds,
		RemoveLabelIds: []string{label},
	}).Do()
	if err != nil {
		log.Printf("failed to batch-delete label %s from %d messages: %s",
			label, len(messages), err)
	}
	// TODO(bzz): move to
	//  gmailutils.ModifyMessagesDelLabel(srv, user, messages, "UNREAD")
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

// TODO(bzz): use a stable sort
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

// paper is a map key, thus aggregation take into account all it's fields.
//
// TODO(bzz): think about aggregation only by the title, as suggested in
// https://github.com/bzz/scholar-alert-digest/issues/12#issuecomment-562820924
type paper struct {
	Title, URL string
	Abstract   abstract
}

type abstract struct {
	Full, FirstLine, RestLines string
}

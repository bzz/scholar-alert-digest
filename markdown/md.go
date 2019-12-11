// Package markdown handles the Markdown generation and it's rendering to HTML.
package markdown

import (
	"io"
	"log"
	"text/template"
	"time"

	"github.com/bzz/scholar-alert-digest/papers"
)

// GenerateMd generates Markdown report for unread emails and the read ones, if any.
func GenerateMd(out io.Writer, newTmplText, oldTmplText string, msgsCnt, titlesCnt int, unread, read map[papers.Paper]int) {
	newMdReport(out, newTmplText, msgsCnt, titlesCnt, unread)
	if read != nil {
		oldMdReport(out, oldTmplText, read)
	}
}

// newMdReport renderes tmplText \w email msg stats (for new, unread papers).
func newMdReport(out io.Writer, tmplText string, msgsCnt, titlesCnt int, agrPapers map[papers.Paper]int) {
	tmpl := template.Must(template.New("unread-papers").Funcs(template.FuncMap{
		"sortedKeys": papers.SortedKeys,
	}).Parse(tmplText))
	err := tmpl.Execute(out, struct {
		Date         string
		UnreadEmails int
		TotalPapers  int
		UniqPapers   int
		Papers       map[papers.Paper]int
	}{
		time.Now().Format(time.RFC3339),
		msgsCnt,
		titlesCnt,
		len(agrPapers),
		agrPapers,
	})
	if err != nil {
		log.Fatalf("template %q execution failed: %s", tmplText, err)
	}
}

// oldMdReport renderes tmplText \wo stats (for old, read papers).
func oldMdReport(out io.Writer, tmplText string, agrPapers map[papers.Paper]int) {
	tmpl := template.Must(template.New("read-papers").Funcs(template.FuncMap{
		"sortedKeys": papers.SortedKeys,
	}).Parse(tmplText))
	err := tmpl.Execute(out, agrPapers)
	if err != nil {
		log.Fatalf("template %q execution failed: %s", tmplText, err)
	}
}

// Package templates hosts all page and report templates
// May be replaced by resources from FS eventually.
package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/bzz/scholar-alert-digest/papers"
	"gitlab.com/golang-commonmark/markdown"
)

var (
	RootLayout = template.Must(template.New("layout").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <link rel="icon" href="http://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/232/page-with-curl_1f4c3.png">
  <base target="_blank">
  <title>{{ template "title" }}</title>
  <style>{{ template "style" }}</style>
</head>
<body>{{ template "body" . }}</body>
</html>
`))

	MdTemplText = `# Google Scholar Alert Digest

**Date**: {{.Date}}
**Unread emails**: {{.UnreadEmails}}
**Paper titles**: {{.TotalPapers}}
**Uniq paper titles**: {{.UniqPapers}}

## New papers
{{ range $title := sortedKeys .Papers }}
   {{ $paper := index $.Papers . }}
 - [{{ $paper.Title }}]({{ $paper.URL }}){{if $paper.Author}}, <i>{{ $paper.Author }}</i>{{end}} {{ template "refs" $paper }}
   {{- if $paper.Abstract.FirstLine }}
   <details>
     <summary>{{ $paper.Abstract.FirstLine }}</summary>
     <div>{{ $paper.Abstract.Rest }}</div>
   </details>
   {{ end }}
{{ end }}
`
	refsMdTemplateText = `
{{ define "refs" -}}
({{ if eq (len .Refs) 0}}{{ .Freq }}{{end}}
{{- if gt (len .Refs) 1 }}{{ .Freq }}: {{ end }}
{{- range $i, $ID := .Refs}}
	{{- if $i}}, {{end}}
	{{- anchorHTML $ID "" $i -}}
{{- end}})
{{- end }}
`

	CompactMdTemplText = `# Google Scholar Alert Digest

**Date**: {{.Date}}
**Unread emails**: {{.UnreadEmails}}
**Paper titles**: {{.TotalPapers}}
**Uniq paper titles**: {{.UniqPapers}}

## New papers
{{ range $title := sortedKeys .Papers }}
   {{ $paper := index $.Papers . }}
 - <details onclick="document.activeElement.blur();">
	 <summary><a href="{{ $paper.URL }}">{{ $paper.Title }}</a>, <i>{{ $paper.Author }}</i> {{ template "refs" $paper }}</summary>
	 <div class="wide">
     {{- if $paper.Abstract.FirstLine }}
	   <div>{{$paper.Abstract.FirstLine}} {{$paper.Abstract.Rest}}</div>
	 {{- end }}
	 </div>
   </details>
{{ end }}
`
	// TODO(bzz): add configurable template for individual li

	ReadMdTemplText = `## Old papers

<details id="archive">
  <summary>Archive</summary>

{{ range $title := sortedKeys . }}
  {{ $paper := index $ . }}
  - [{{ $paper.Title }}]({{ $paper.URL }})
    {{- if $paper.Abstract.FirstLine }}
    <details>
      <summary>{{$paper.Abstract.FirstLine}}</summary>{{$paper.Abstract.Rest}}
    </details>
    {{ end }}
{{ end }}
</details>
`

	CompatStyle = `
ul { list-style-type: none; margin: 0; padding: 0 0 0 20px; }
#archive>ul {list-style-type: circle; }
.wide { max-width:60%; margin-left: 1em; padding: 0.2em 0 0.5em 0; }
`
)

// Renderer renders papers in one of the supported output formats: Markdown/HTML/JSON/JSONL.
type Renderer interface {
	Render(out io.Writer, st *papers.Stats, unread, read papers.AggPapers)
}

// JSONRenderer outputs JSON/JSONL formats.
type JSONRenderer struct {
	render func(io.Writer, *papers.Stats, papers.AggPapers, papers.AggPapers)
}

// Render papers in JSON/JSONL.
func (r *JSONRenderer) Render(out io.Writer, st *papers.Stats, unread, read papers.AggPapers) {
	r.render(out, st, unread, read)
}

// NewJSONRenderer factory for Renderer in JSON format.
func NewJSONRenderer() Renderer {
	return &JSONRenderer{
		render: func(out io.Writer, st *papers.Stats, unread, read papers.AggPapers) {
			log.Printf("formatting gmail messages in JSON")
			t := []*papers.Paper{}
			all := map[string]interface{}{
				"read": nil, "unread": nil, "stats": st,
			}

			for _, title := range papers.SortedKeys(read) {
				t = append(t, read[title])
			}
			all["read"] = t
			t = nil

			for _, title := range papers.SortedKeys(unread) {
				t = append(t, unread[title])
			}
			all["unread"] = t
			t = nil

			encoder := json.NewEncoder(out)
			encoder.Encode(all)
		},
	}
}

// NewJSONLRenderer factory for Renderer in JSONL format.
func NewJSONLRenderer() Renderer {
	return &JSONRenderer{
		render: func(out io.Writer, st *papers.Stats, unread, read papers.AggPapers) {
			log.Print("formatting gmail messages in JSONL")
			encoder := json.NewEncoder(out)
			for _, title := range papers.SortedKeys(unread) {
				encoder.Encode(unread[title])
			}
			for _, title := range papers.SortedKeys(read) {
				encoder.Encode(read[title])
			}
		},
	}
}

// MarkdownRenderer outputs Markdown.
type MarkdownRenderer struct {
	layout     *template.Template
	template   string
	oldTempate string
}

func NewMarkdownRenderer(templateText, oldTemplateText string) Renderer {
	return &MarkdownRenderer{
		template.New("papers").Funcs(template.FuncMap{
			"sortedKeys": papers.SortedKeys,
			"anchorHTML": func(ID, title string, i int) template.HTML {
				if title == "" {
					title = strconv.Itoa(i + 1)
				}
				// Needed for re-use of refsMdTemplate between -compact and normal Md:
				//  * in -compact, markdown syntax for links will not be rendered inside <summary>
				//  * html/template escape HTML strings \wo template.HTML
				return template.HTML(
					fmt.Sprintf(
						"<a target='_blank' href='https://mail.google.com/mail/#inbox/%s'>%s</a>",
						ID, title,
					),
				)
			},
		}),
		templateText,
		oldTemplateText,
	}
}

func (r *MarkdownRenderer) Render(out io.Writer, st *papers.Stats, unread, read papers.AggPapers) {
	r.newMdReport(out, st, unread)
	if read != nil {
		r.oldMdReport(out, read)
	}
}

// newMdReport renderes tmplText \w email msg stats (for new, unread papers).
func (r *MarkdownRenderer) newMdReport(out io.Writer, st *papers.Stats, agrPapers papers.AggPapers) {
	layout := template.Must(r.layout.Clone())
	tmpl := template.Must(layout.Parse(r.template))
	tmpl = template.Must(tmpl.Parse(refsMdTemplateText))
	err := tmpl.Execute(out, struct {
		Date         string
		UnreadEmails int
		TotalPapers  int
		UniqPapers   int
		Papers       papers.AggPapers
	}{
		time.Now().Format(time.RFC3339),
		st.Msgs,
		st.Titles,
		len(agrPapers),
		agrPapers,
	})
	if err != nil {
		log.Fatalf("template %q execution failed: %s", r.template, err)
	}
}

// oldMdReport renderes tmplText \wo stats (for old, read papers).
func (r *MarkdownRenderer) oldMdReport(out io.Writer, agrPapers papers.AggPapers) {
	layout := template.Must(r.layout.Clone())
	tmpl := template.Must(layout.Parse(r.oldTempate))
	err := tmpl.Execute(out, agrPapers)
	if err != nil {
		log.Fatalf("template %q execution failed: %s", r.oldTempate, err)
	}
}

// HTMLRenderer outputs HTML from template in Markdown.
type HTMLRenderer struct {
	Renderer
	layout *template.Template
	style  string
}

func NewHTMLRenderer(templateText, style string) Renderer {
	return &HTMLRenderer{NewMarkdownRenderer(templateText, ReadMdTemplText), RootLayout, style}
}

func (r *HTMLRenderer) Render(out io.Writer, st *papers.Stats, unread, read papers.AggPapers) {
	var mdBuf bytes.Buffer
	r.Renderer.Render(&mdBuf, st, unread, read)

	var htmlBuf bytes.Buffer
	md := markdown.New(markdown.XHTMLOutput(true), markdown.HTML(true))
	md.Render(&htmlBuf, mdBuf.Bytes())

	// rootLayout requires 3 sub-templates
	title := `{{ define "title" }}scholar alert digest{{ end }}`
	style := fmt.Sprintf(`{{ define "style" }}%s{{ end }}`, r.style)
	body := fmt.Sprintf(`{{ define "body" }}%s{{ end }}`, htmlBuf.String())

	// TODO(bzz): move tmpl construction out of .Render(), so there is either:
	// - only one .Clone() + .Parese() for dynamic "body" template, generated from MD
	// - or change "body" template to recive HTML generated from MD as a data (better)
	tmpl := template.Must(r.layout.Clone())
	tmpl = template.Must(tmpl.Parse(title))
	tmpl = template.Must(tmpl.Parse(style))
	tmpl = template.Must(tmpl.Parse(body))
	err := tmpl.Execute(out, nil)
	if err != nil {
		log.Fatalf("template execution failed: %s", err)
	}
}

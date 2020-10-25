package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/bzz/scholar-alert-digest/gmailutils"
	"github.com/bzz/scholar-alert-digest/gmailutils/token"
	"github.com/bzz/scholar-alert-digest/papers"
	"github.com/bzz/scholar-alert-digest/templates"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const user = "me"

var ( // templates
	chooseLabelsForm = `
{{ define "title" }}Chose a label{{ end }}
{{ define "style" }}{{ end }}
{{ define "body" }}
<p>Please, chosse a Gmail label to aggregate:</p>
<form action="/labels" method="POST">
{{ range . }}
    <div>
      <input type="radio" id="{{.}}" name="label" value="{{.}}">
      <label for="{{.}}">{{.}}</label>
	</div>
{{ end }}

  <input type="submit" value="Select Label"/>
</form>
{{ end }}
`

	newMdTemplText = `# Google Scholar Alert Digest

**Date**: {{.Date}}
**Unread emails**: {{.UnreadEmails}}
**Paper titles**: {{.TotalPapers}}
**Uniq paper titles**: {{.UniqPapers}}

## New papers
{{ range $paper := sortedKeys .Papers }}
 - [{{ .Title }}]({{ .URL }}) ({{index $.Papers .}})
   {{- if .Abstract.FirstLine }}
   <details>
     <summary>{{.Abstract.FirstLine}}</summary>{{.Abstract.Rest}}
   </details>
   {{ end }}
{{ end }}
`
)

var ( // configuration
	addr      = "localhost:8080"
	concurReq = 10
	oauthCfg  = &oauth2.Config{
		// from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		ClientID:     os.Getenv("SAD_GOOGLE_ID"),
		ClientSecret: os.Getenv("SAD_GOOGLE_SECRET"),
		RedirectURL:  "http://localhost:8080/login/authorized",
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailReadonlyScope},
	}
)

var ( // CLI
	compact = flag.Bool("compact", false, "output report in compact format (>100 papers)")
	test    = flag.Bool("test", false, "read emails from ./fixtures/* instead of real Gmail")
	cors    = flag.Bool("cors", false, "enable CORS for all endpoints (usefull for development)")
	// TODO(bzz): add -read support + equivalent per-user config option (cookies)
)

var htmlRn, jsonRn templates.Renderer

func main() {
	flag.Parse()

	templateText, style := templates.MdTemplText, ""
	if *compact {
		templateText, style = templates.CompactMdTemplText, templates.CompatStyle
	}
	htmlRn = templates.NewHTMLRenderer(templateText, style)
	jsonRn = templates.NewJSONRenderer()

	// TODO(bzz):
	//  - configure the log level, to include requests in debug
	//  - add default req timeouts + throttling, to prevent abuse

	log.Printf("starting the web server at http://%s", addr)
	defer log.Printf("stoping the web server")

	r := chi.NewRouter()
	r.Use(tokenAndLabelCookiesCtx)
	// r.Use(middleware.Logger)

	r.Get("/", handleRoot)
	r.Get("/labels", handleLabelsRead)
	r.Post("/labels", handleLabelsWrite)
	r.Get("/login", handleLogin)
	r.Get("/login/authorized", handleAuth)

	r.Route("/json", func(j chi.Router) {
		j.Use(render.SetContentType(render.ContentTypeJSON))
		if !*test {
			j.Use(tokenCtx) // read cookies, else - advise authCode URL to frontend
			j.Use(labelCtx) // read cookies, else - advise /labels URL to frontend (only if not on /labels)
		}
		if *cors {
			j.Use(middlewareCORS)
		}

		j.Get("/labels", listLabels)
		j.Get("/messages", listMessages)
		// j.Get("/papers", handlePaperssJSON)
	})

	http.ListenAndServe(addr, r)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// get token, stored in context by middleware (from cookies)
	tok, authorized := token.FromContext(r.Context())
	if !authorized && !*test { // TODO(bzz): move this to middleware
		log.Printf("Redirecting to /login as three is no session")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	gmailLabel, hasLabel := token.LabelFromContext(r.Context())
	if !hasLabel && !*test {
		log.Printf("Redirecting to /labels as there is no label")
		http.Redirect(w, r, "/labels", http.StatusFound)
		return
	}

	// find and fetch email messages
	var rMsgs, urMsgs []*gmail.Message
	if !*test { // TODO(bzz): refactor, replace \w polymorphism though interface for fetching messages
		var err error
		srv, _ := gmail.New(oauthCfg.Client(r.Context(), tok)) // ignore err as client != nil
		query := fmt.Sprintf("label:%s is:unread", gmailLabel)
		urMsgs, err = gmailutils.FetchConcurent(r.Context(), srv, user, query, concurReq)
		if err != nil {
			// TODO(bzz): token expiration looks ugly here and must be handled elsewhere
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(err.Error()))
			return
		}
	} else {
		urMsgs = gmailutils.ReadFixturesJSON("./fixtures/unread.json")
		rMsgs = gmailutils.ReadFixturesJSON("./fixtures/read.json")
	}

	// aggregate
	urStats, urTitles := papers.ExtractAndAggPapersFromMsgs(urMsgs, true, true)
	if urStats.Errs != 0 {
		log.Printf("%d errors found, extracting the papers", urStats.Errs)
	}

	rStats, rTitles := papers.ExtractAndAggPapersFromMsgs(rMsgs, true, true)
	if rStats.Errs != 0 {
		log.Printf("%d errors found, extracting the papers", rStats.Errs)
	}

	// render
	if _, ok := r.URL.Query()["json"]; ok {
		w.Header().Set("Content-Type", "application/json")
		jsonRn.Render(w, urStats, urTitles, rTitles)
	} else {
		htmlRn.Render(w, urStats, urTitles, nil)
	}
}

func handleLabelsRead(w http.ResponseWriter, r *http.Request) {
	tok, authorized := token.FromContext(r.Context())
	if !authorized { // TODO(bzz): move this to middleware
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	fetchLabelsAndServeForm(w, r, tok)
}

func handleLabelsWrite(w http.ResponseWriter, r *http.Request) {
	saveLabelToCookies(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func fetchLabelsAndServeForm(w http.ResponseWriter, r *http.Request, tok *oauth2.Token) {
	client := oauthCfg.Client(r.Context(), tok)
	labelsResp, err := gmailutils.FetchLabels(r.Context(), client)
	if err != nil {
		log.Printf("Unable to retrieve all labels: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var labels []string // user labels, sorted
	for _, l := range labelsResp.Labels {
		if l.Type == "system" {
			continue
		}
		labels = append(labels, l.Name)
	}
	sort.Strings(labels)

	// render combination of the nested templates
	tmpl := template.Must(templates.RootLayout.Clone())
	tmpl = template.Must(tmpl.Parse(chooseLabelsForm))
	err = tmpl.Execute(w, labels)
	if err != nil {
		log.Printf("Failed to render a template: %v", err)
	}
}

func saveLabelToCookies(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil { // url query part is not valid
		log.Printf("Unable to parse query string: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	humanLabel := r.FormValue("label")
	label := gmailutils.FormatAsID(humanLabel)

	cookie := token.NewLabelCookie(label)
	log.Printf("Saving new cookie: %s", cookie.String())
	http.SetCookie(w, cookie)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// the URL which shows the Google Auth page to the user
	url := oauthCfg.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusFound)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	// get the code from URL
	err := r.ParseForm()
	if err != nil { // url query part is not valid
		log.Printf("Unable to parse query string: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// exchange the received code for a bearer token
	code := r.FormValue("code")
	// FIXME(bzz): validate that r.FormValue("state") is the same as in AuthCodeURL()
	tok, err := oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Unable to exchange the code %q for token: %v", code, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// save token in the session cookie
	cookie := token.NewSessionCookie(tok)
	log.Printf("Saving new cookie: %s", cookie.String())
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}

// tokenAndLabelCookiesCtx reads token from request cookie and saves it to the Context.
func tokenAndLabelCookiesCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, "-", r.RequestURI /*, r.Cookies()*/) // TODO(bzz): make cookies debug level only
		ctx := token.NewSessionContext(r.Context(), r.Cookies())
		ctx = token.NewLabelContext(ctx, r.Cookies())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//// JSON

// middlewareCORS enables CORS.
func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const (
	tokenKey contextKey = "token"
	labelKey contextKey = "label"
)

func listLabels(w http.ResponseWriter, r *http.Request) {
	tok := r.Context().Value(tokenKey).(*oauth2.Token)
	client := oauthCfg.Client(r.Context(), tok)
	labelsResp, err := gmailutils.FetchLabels(r.Context(), client)
	if err != nil {
		// TODO pass Error to frontend
		// render.Render(w, r, ErrNotFound)
		log.Printf("Unable to retrieve all labels: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var labels []string // user labels, sorted
	for _, l := range labelsResp.Labels {
		if l.Type == "system" {
			continue
		}
		labels = append(labels, l.Name)
	}
	sort.Strings(labels)

	// TODO
	json.NewEncoder(w).Encode(map[string]interface{}{"labels": labels})
	// render.RenderList(w, r, labels)
}

func listMessages(w http.ResponseWriter, r *http.Request) {

}

type ErrResponse struct {
	Err            string `json:",omitempty"` // low-level runtime error
	HTTPStatusCode int    `json:"-"`          // http response status code

	StatusText string `json:"status"` // user-level status message
	Redirect   string `json:",omitempty"`
}

func tokenCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// read from cookies
		var tokenCookie *http.Cookie
		for _, c := range r.Cookies() {
			if c.Name == string(tokenKey) {
				tokenCookie = c
				break
			}
		}

		if tokenCookie == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrResponse{
				StatusText: "Not authorized",
				Redirect:   oauthCfg.AuthCodeURL("scholar"),
			})
			return
		}

		// decode 	64
		data, err := base64.StdEncoding.DecodeString(tokenCookie.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(ErrResponse{
				Err:        err.Error(),
				StatusText: "Unable to decode base64 cookie: " + string(tokenKey),
			})
			return
		}

		// decode JSON
		tok := &oauth2.Token{}
		err = json.NewDecoder(bytes.NewReader(data)).Decode(tok)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(ErrResponse{
				Err:        err.Error(),
				StatusText: "Unable to decode JSON token: " + string(data),
			})
			return
		}

		ctx = context.WithValue(ctx, tokenKey, tok)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func labelCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// TODO
		// read from cookies, decode base64
		var val string

		ctx = context.WithValue(ctx, labelKey, val)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

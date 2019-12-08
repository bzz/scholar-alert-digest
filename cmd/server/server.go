package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/bzz/scholar-alert-digest/gmailutils"
	"github.com/bzz/scholar-alert-digest/gmailutils/token"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var ( // templates
	layout = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{ template "title" }}</title>
  <link rel="icon" href="http://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/232/page-with-curl_1f4c3.png">
</head>
<body>{{ template "body" . }}</body>
</html>
`
	chooseLabelsForm = `
{{ define "title"}}Chose a label{{ end }}
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
)

var ( // configuration
	addr     = "localhost:8080"
	oauthCfg = &oauth2.Config{
		// from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		ClientID:     os.Getenv("SAL_GOOGLE_ID"),
		ClientSecret: os.Getenv("SAL_GOOGLE_SECRET"),
		RedirectURL:  "http://localhost:8080/login/authorized",
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailReadonlyScope},
	}
)

func main() {
	// TODO(bzz):
	//  - configure the log level, to include requests in debug
	//  - add default req timeouts + throttling, to prevent abuse

	log.Printf("starting the web server at http://%s", addr)
	defer log.Printf("stoping the web server")

	// TODO(bzz): migrate to chi-router
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/labels", handleLabels)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/login/authorized", handleAuth)

	http.ListenAndServe(addr, sessionMiddleware(mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// get token, stored in context by middleware (from cookies)
	tok, authorized := token.FromContext(r.Context())
	if !authorized { // TODO(bzz): move this to middleware
		log.Printf("Redirecting to /login as three is no session")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	gmailLabel, hasLabel := token.LabelFromContext(r.Context())
	if !hasLabel {
		log.Printf("Redirecting to /labels as there is no label")
		http.Redirect(w, r, "/labels", http.StatusFound)
		return
	}

	// fetch messages
	srv, _ := gmail.New(oauthCfg.Client(r.Context(), tok)) // ignore as client != nil
	msgs, err := gmailutils.Fetch(r.Context(), srv, "me", fmt.Sprintf("label:%s is:unread", gmailLabel))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO(bzz):
	//  add spniner? fetching takes ~20 sec for me
	//  aggregate msgs
	//  render HTML template

	// just print, for now
	json.NewEncoder(w).Encode(msgs) // FIXME(bzz): handle JSON encoding failures
}

func handleLabels(w http.ResponseWriter, r *http.Request) {
	tok, authorized := token.FromContext(r.Context())
	if !authorized { // TODO(bzz): move this to middleware
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		fetchLabelsAndServeForm(w, r, tok)
	case http.MethodPost:
		saveLabelToCookies(w, r)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func fetchLabelsAndServeForm(w http.ResponseWriter, r *http.Request, tok *oauth2.Token) {
	labelsResp, err := gmailutils.FetchLabels(r.Context(), oauthCfg, tok)
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

	tmpl := template.Must( // render combination of the nested templates
		template.Must(
			template.New("choose-label").Parse(layout)).
			Parse(chooseLabelsForm))
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

// sessionMiddleware reads token from session cookie, saves it into the Context.
func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, "-", r.RequestURI /*, r.Cookies()*/) // TODO(bzz): make cookies debug level only
		ctx := token.NewSessionContext(r.Context(), r.Cookies())
		ctx = token.NewLabelContext(ctx, r.Cookies())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

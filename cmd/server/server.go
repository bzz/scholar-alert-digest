package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bzz/scholar-alert-digest/gmailutils"
	"github.com/bzz/scholar-alert-digest/gmailutils/token"
	js "github.com/bzz/scholar-alert-digest/json"
	"github.com/bzz/scholar-alert-digest/papers"
	"github.com/bzz/scholar-alert-digest/templates"
	"github.com/rs/cors"

	"github.com/go-chi/chi"
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
	dev     = flag.Bool("dev", false, "development mode where /login/auth redirects to :9000 and CORS is enabled")
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

	corsOptions := cors.Options{
		AllowedOrigins:   []string{"http://localhost:9000"},
		AllowCredentials: true,
		Debug:            true,
	}

	r.Get("/", handleRoot)
	r.Get("/labels", handleLabelsRead)
	r.Post("/labels", handleLabelsWrite)
	r.Get("/login", handleLogin)
	r.Get("/login/authorized", handleAuth)

	if !*dev { // static
		workDir, _ := os.Getwd()
		filesDir := http.Dir(filepath.Join(workDir, "frontend", "dist"))
		FileServer(r, "/static", filesDir)
	}

	r.Route("/json", func(j chi.Router) {
		j.Use(setContentType("application/json"))
		if *dev {
			j.Use(cors.New(corsOptions).Handler)
		}
		if !*test {
			j.Use(tokenCtx)
		}

		j.Get("/labels", listLabels)
		j.With(labelCtx).Post("/messages", listMessages)
		// j.Get("/papers", listPapers)
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
		urMsgs = gmailutils.ReadMsgFixturesJSON("./fixtures/unread.json")
		rMsgs = gmailutils.ReadMsgFixturesJSON("./fixtures/read.json")
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
	var gmLabels []*gmail.Label
	if !*test {
		tok, authorized := token.FromContext(r.Context())
		if !authorized { // TODO(bzz): move this to middleware
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		client := oauthCfg.Client(r.Context(), tok)
		labelsResp, err := gmailutils.FetchLabels(r.Context(), client)
		if err != nil {
			log.Printf("Unable to retrieve all labels: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gmLabels = labelsResp.Labels
	} else {
		gmLabels = gmailutils.ReadLblFixturesJSON("./fixtures/labels.json")
	}

	var labels []string // user labels, sorted
	for _, l := range gmLabels {
		if l.Type == "system" {
			continue
		}
		labels = append(labels, l.Name)
	}
	sort.Strings(labels)

	// render combination of the nested templates
	tmpl := template.Must(templates.RootLayout.Clone())
	tmpl = template.Must(tmpl.Parse(chooseLabelsForm))
	err := tmpl.Execute(w, labels)
	if err != nil {
		log.Printf("Failed to render a template: %v", err)
	}
}

func handleLabelsWrite(w http.ResponseWriter, r *http.Request) {
	saveLabelToCookies(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
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

	toURL := "/"
	if *dev {
		toURL = "//localhost:9000"
	}
	http.Redirect(w, r, toURL, http.StatusMovedPermanently)
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

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

//// JSON

func setContentType(contentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type contextKey string

func (k contextKey) readFrom(cookies []*http.Cookie) *http.Cookie {
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == string(k) {
			cookie = c
			break
		}
	}
	return cookie
}

const (
	tokenKey contextKey = "token"
	labelKey contextKey = "label"
)

func listLabels(w http.ResponseWriter, r *http.Request) {
	var gmLabels []*gmail.Label
	if !*test {
		tok := r.Context().Value(tokenKey).(*oauth2.Token)
		client := oauthCfg.Client(r.Context(), tok)
		labelsResp, err := gmailutils.FetchLabels(r.Context(), client)
		if err != nil {
			js.ErrNotFound(w, err, "Unable to retrieve labels from Gmail")
			return
		}
		gmLabels = labelsResp.Labels
	} else {
		gmLabels = gmailutils.ReadLblFixturesJSON("./fixtures/labels.json")
	}

	var labels []string // user labels, sorted
	for _, l := range gmLabels {
		if l.Type == "system" {
			continue
		}
		labels = append(labels, gmailutils.FormatAsID(l.Name))
	}
	sort.Strings(labels)

	json.NewEncoder(w).Encode(map[string]interface{}{"labels": labels})
	// TODO: better rendering
	// render.RenderList(w, r, labels)
}

func listMessages(w http.ResponseWriter, r *http.Request) {
	label := r.Context().Value(labelKey).(string)

	var err error
	var urMsgs, rMsgs []*gmail.Message
	if !*test { // TODO(bzz): refactor, replace \w polymorphism though interface for fetching messages
		tok := r.Context().Value(tokenKey).(*oauth2.Token)
		srv, _ := gmail.New(oauthCfg.Client(r.Context(), tok)) // ignore err as client != nil
		query := fmt.Sprintf("label:%s is:unread", label)
		urMsgs, err = gmailutils.FetchConcurent(r.Context(), srv, user, query, concurReq)
		if err != nil {
			js.ErrFailedDependency(w, err, "failed to fetch messages from Gmail")
			return
		}
	} else {
		urMsgs = gmailutils.ReadMsgFixturesJSON("./fixtures/unread.json")
		rMsgs = gmailutils.ReadMsgFixturesJSON("./fixtures/read.json")
	}

	// aggregate
	urStats, urTitles := papers.ExtractAndAggPapersFromMsgs(urMsgs, true, true)
	if urStats.Errs != 0 {
		log.Printf("%d errors found, extracting the papers", urStats.Errs)
	}

	rStats, rTitles := papers.ExtractAndAggPapersFromMsgs(rMsgs, true, true)
	if rStats.Errs != 0 {
		log.Printf("%d errors found, extracting the papers", urStats.Errs)
	}

	jsonRn.Render(w, urStats, urTitles, rTitles)
}

func tokenCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenCookie := tokenKey.readFrom(r.Cookies())
		if tokenCookie == nil {
			js.ErrUnauthorized(w, oauthCfg.AuthCodeURL("scholar"))
			return
		}

		data, err := base64.StdEncoding.DecodeString(tokenCookie.Value)
		if err != nil {
			js.ErrUnprocessable(w, err, "Unable to decode base64 cookie: "+string(tokenKey))
			return
		}

		tok := &oauth2.Token{}
		err = json.NewDecoder(bytes.NewReader(data)).Decode(tok)
		if err != nil {
			js.ErrUnprocessable(w, err, "Unable to decode JSON token: "+string(data))
			return
		}

		ctx = context.WithValue(ctx, tokenKey, tok)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func labelCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			js.ErrUnprocessable(w, err, "Failed to read request body")
			return
		}

		req := map[string]string{}
		err = json.Unmarshal(data, &req)
		if err != nil {
			js.ErrUnprocessable(w, err, "Unable to decode JSON label: "+string(data))
			return
		}

		if label, ok := req[string(labelKey)]; ok {
			ctx = context.WithValue(ctx, labelKey, label)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/bzz/scholar-alert-digest/gmailutils"
	"github.com/bzz/scholar-alert-digest/gmailutils/token"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/login/authorized", handleAuth)

	http.ListenAndServe(addr, sessionMiddleware(mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// get token, stored in context by middleware (from cookies)
	tok, authorized := token.FromContext(r.Context())
	if !authorized { // TODO(bzz): move this to middleware
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not logged in: go to /login\n"))
	}

	// TOOD(bzz): if the label is known, fetch messages instead
	// gmailutils.Fetch(srv, "me", fmt.Sprintf("label:%s is:unread", gmailLabel))
	labels, err := gmailutils.FetchLabels(r.Context(), oauthCfg, tok)
	if err != nil {
		log.Printf("Unable to retrieve all labels: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: render a template \w POST form to label selection
	data, err := labels.MarshalJSON()
	if err != nil {
		log.Printf("Failed to encode labels in JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(data)
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
		log.Println(r.Method, "-", r.RequestURI /*, r.Cookies()*/)
		ctx := token.NewSessionContext(r.Context(), r.Cookies())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

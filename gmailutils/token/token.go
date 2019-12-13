/**
 * Copyright Google Inc.
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

// Package token provides utilities for caching Gmail auth token.
package token

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/oauth2"
)

// FromWeb request a token from the web, then returns the retrieved token.
func FromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprintf(os.Stderr, "Open this link in the browser, then past the "+
		"authorization code: \n%v\n", authURL)

	_ = openURL(authURL) // ignore error as manual instuctions already provided

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// openURL opens a browser window to the specified location.
// This code originally appeared at:
//   http://stackoverflow.com/questions/10377243/how-can-i-launch-a-process-that-is-not-a-file-in-go
func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:4001/").Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Cannot open URL %s on this platform", url)
	}
	return err
}

// FromFile retrieves a token from a local file.
func FromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// TODO(bzz): move to session.go, cookies/context are together due to sharing the Key
// contextKey is unexported type to prevent collisions with context keys.
type contextKey string

const (
	sessionKey contextKey = "session"
	labelKey   contextKey = "label"
)

// FromContext returnes the token, saved from the cookies, if any.
func FromContext(ctx context.Context) (*oauth2.Token, bool) {
	token := ctx.Value(sessionKey)
	if token == nil { // not authorized
		return nil, false
	}

	var tok oauth2.Token
	tokenStr := token.(string)
	err := json.NewDecoder(strings.NewReader(tokenStr)).Decode(&tok)
	if err != nil {
		log.Printf("Unable to decode JSON cookie k:%s v:%s, %v", sessionKey, tokenStr, err)
		return nil, true
	}
	return &tok, true
}

// LabelFromContext returnes the label, saved from the cookies, if any.
func LabelFromContext(ctx context.Context) (string, bool) {
	l := ctx.Value(labelKey)
	if l == nil {
		return "", false
	}

	return l.(string), true
}

// NewSessionCookie returns a new cookie with the token set.
func NewSessionCookie(token *oauth2.Token) *http.Cookie {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(token) // FIXME(bzz): handle JSON encoding failures
	sessionVal := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &http.Cookie{
		Name:     string(sessionKey),
		Value:    sessionVal,
		Path:     "/",
		Expires:  token.Expiry,
		HttpOnly: true,
	}
}

// NewLabelCookie returns a new cookie with the token set.
func NewLabelCookie(label string) *http.Cookie {
	labelVal := base64.StdEncoding.EncodeToString([]byte(label))

	return &http.Cookie{
		Name:  string(labelKey),
		Value: labelVal,
		Path:  "/",
	}
}

// NewSessionContext reads session cookie, returnes context with the token set to it, if any.
func NewSessionContext(parent context.Context, cookies []*http.Cookie) context.Context {
	return newContextWith(parent, cookies, sessionKey)
}

// NewLabelContext reads label cookie, returnes context with the label set to it, if any.
func NewLabelContext(parent context.Context, cookies []*http.Cookie) context.Context {
	return newContextWith(parent, cookies, labelKey)
}

func newContextWith(parent context.Context, cookies []*http.Cookie, key contextKey) context.Context {
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == string(key) {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		return parent
	}

	sessionVal, err := base64.StdEncoding.DecodeString(sessionCookie.Value)
	if err != nil {
		log.Printf("Unable to decode base64 %s cookie: %v", string(key), err)
		return parent
	}
	return context.WithValue(parent, key, string(sessionVal))
}

// Save saves the token to a file path.
func Save(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

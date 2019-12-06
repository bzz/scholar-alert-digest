/**
 * Copyright Google Inc.
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

// Package gmailutils provides helpers for Gmail API.
package gmailutils

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bzz/scholar-alert-digest/gmailutils/token"

	"github.com/cheggaaa/pb/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Instructions are user manual for OAuth app configuration from Gmail.
const Instructions = `Please follow https://developers.google.com/gmail/api/quickstart/go#step_1_turn_on_the
in oreder to:
 - create a new "Quickstart" API project under your account
 - enable GMail API on it
 - download OAuth 2.0 credentials
`

// NewClient a client configured with OAuth using 'credentials.json' and a 'token.json'.
func NewClient(needWriteAccess bool) *http.Client {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v\n%s", err, Instructions)
	}

	// If modifying these scopes, delete your previously saved token.json.
	scopes := []string{gmail.GmailReadonlyScope}
	token := "token.json"
	if needWriteAccess {
		scopes = append(scopes, gmail.GmailModifyScope)
		token = "token_rw.json"
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return getClient(config, token)
}

// Retrieve an OAuth token, saves it, then returns a pre-configured client.
func getClient(config *oauth2.Config, tokFile string) *http.Client {
	// The tokFile stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tok, err := token.FromFile(tokFile)
	if err != nil {
		tok = token.FromWeb(config)
		token.Save(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// FetchLabels fetches the list of labels, as returned by Gmail.
func FetchLabels(ctx context.Context, oauthCfg *oauth2.Config, token *oauth2.Token) (
	*gmail.ListLabelsResponse, error) { // TODO(bzz): extract all args to a struct and make it a method
	// get an authorized Gmail API client
	client := oauthCfg.Client(ctx, token)
	srv, err := gmail.New(client)
	if err != nil {
		return nil, err
	}

	// TODO(bzz): handle token expiration (by cookie expiration? or set refresh token?)
	// Unable to retrieve all labels: Get https://www.googleapis.com/gmail/v1/users/me/labels?alt=json&prettyPrint=false: oauth2: token expired and refresh token is not set

	// fetch from Gmail
	lablesResp, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return nil, err
	}
	return lablesResp, nil
}

// PrintAllLabels prints all labels for a given user.
func PrintAllLabels(srv *gmail.Service, user string) {
	log.Printf("Listing all Gmail labels")
	lablesResp, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve all labels: %v", err)
	}

	log.Printf("%d labels found", len(lablesResp.Labels))
	for _, label := range lablesResp.Labels {
		fmt.Printf("%s\n", strings.ToLower(strings.ReplaceAll(label.Name, " ", "-")))
	}
}

// Fetch fetches all messages matching a given query from the Gmail.
func Fetch(srv *gmail.Service, user, query string) []*gmail.Message {
	start := time.Now()
	msgs := QueryMessages(srv, user, query)
	log.Printf("%d messages found under %q (took %.0f sec)", len(msgs), query, time.Since(start).Seconds())
	return msgs
}

// QueryMessages returns the all messages, matching a query for a given user.
func QueryMessages(srv *gmail.Service, user, query string) []*gmail.Message {
	var messages []*gmail.Message
	page := 0 // iterate pages
	err := srv.Users.Messages.List(user).Q(query).Pages(context.TODO(), func(mr *gmail.ListMessagesResponse) error {
		log.Printf("page %d: found %d messages, fetching ...", page, len(mr.Messages)) // TODO(bzz): debug level only

		bar := pb.Full.Start(len(mr.Messages))
		bar.SetMaxWidth(100)
		for _, m := range mr.Messages {
			bar.Increment()
			msg, err := srv.Users.Messages.Get(user, m.Id).Do()
			if err != nil {
				return err
			}

			messages = append(messages, msg)
		}

		bar.Finish()
		log.Printf("page %d: %d messaged fetched", page, len(mr.Messages)) // TODO(bzz): debug level only
		page++
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to retrieve messages with the query %q, page %d: %v", query, page, err)
	}

	return messages
}

// Subject returns the Subject header of a message
func Subject(m *gmail.MessagePart) string {
	if m == nil {
		return ""
	}

	for _, h := range m.Headers {
		if h.Name == "Subject" {
			return h.Value
		}
	}
	return ""
}

// MessageTextBody returns the text (if any) of a given message ID
func MessageTextBody(m *gmail.Message) ([]byte, error) {
	body, _, err := recursiveDecodeParts(m.Payload, "text/html")
	if body == nil {
		return nil, errors.New("no message payload")
	}
	return body, err
}

func recursiveDecodeParts(part *gmail.MessagePart, mimeType string) ([]byte, string, error) {
	if part == nil || part.Body == nil {
		return nil, "", nil
	}

	var gotError error
	for _, p := range part.Parts {
		b, m, err := recursiveDecodeParts(p, mimeType)
		if b != nil || m != "" {
			return b, m, err
		}
		if err != nil {
			gotError = err
		}
	}

	switch {
	case strings.HasPrefix(part.MimeType, mimeType):
		if part.Body.AttachmentId != "" {
			return nil, part.Body.AttachmentId, nil
		}
		b, err := base64.StdEncoding.DecodeString(part.Body.Data)
		if err != nil {
			b, err = base64.URLEncoding.DecodeString(part.Body.Data)
		}
		return b, "", err
	}
	return nil, "", gotError
}

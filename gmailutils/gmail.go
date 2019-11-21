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

	"github.com/bzz/scholar-alert-digest/gmailutils/token"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// NewClient a client configured with OAuth using 'credentials.json' and a 'token.json'.
func NewClient() *http.Client {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return getClient(config)
}

// Retrieve an OAuth token, saves it, then returns a pre-configured client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := token.FromFile(tokFile)
	if err != nil {
		tok = token.FromWeb(config)
		token.Save(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// UnreadMessagesInLabel returns unread messages under a given lable.
func UnreadMessagesInLabel(srv *gmail.Service, user, labelName string) []*gmail.Message {
	log.Printf("Searching for all unread messages under Gmail label %q", labelName)
	query := fmt.Sprintf("label:%s is:unread", labelName)
	return queryMessages(srv, user, query)
}

// queryMessages returns all messages matching a query for a given user.
func queryMessages(srv *gmail.Service, user, query string) []*gmail.Message {
	var messages []*gmail.Message
	page := 0 // iterate pages
	err := srv.Users.Messages.List(user).Q(query).Pages(context.TODO(), func(rm *gmail.ListMessagesResponse) error {
		for _, m := range rm.Messages {
			msg, err := srv.Users.Messages.Get(user, m.Id).Do()
			if err != nil {
				return err
			}

			messages = append(messages, msg)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Unable to retrieve messages with query %q, page %d: %v", query, page, err)
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

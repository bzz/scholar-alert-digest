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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

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
	labelsResp, err := srv.Users.Labels.List("me").Do()
	if err != nil {
		return nil, err
	}
	return labelsResp, nil
}

// PrintAllLabels prints all labels for a given user.
func PrintAllLabels(srv *gmail.Service, user string) {
	log.Printf("Listing all Gmail labels")
	labelsResp, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve all labels: %v", err)
	}

	log.Printf("%d labels found", len(labelsResp.Labels))
	for _, l := range labelsResp.Labels {
		fmt.Println(FormatAsID(l.Name))
	}
}

// FetchConcurent fetches matching messages for a given query in paralle from the Gmail.
// It is blocking, but doing N concurrent fetche requests.
// TODO(bzz): make it a method on the struct, that holds srv instance.
func FetchConcurent(ctx context.Context, srv *gmail.Service, user, query string, concurentReq int) ([]*gmail.Message, error) {
	log.Printf("searching and fetching messages from Gmail: %q", query)
	start := time.Now()
	msgs, err := searchAndFetchConcurent(ctx, srv, user, query, concurentReq)
	if err != nil {
		return nil, err
	}
	log.Printf("%d messages found&fetched with (took %.0f sec)", len(msgs), time.Since(start).Seconds())
	return msgs, nil
}

// TODO(bzz): make it a method on the struct, that holds srv instance.
func searchAndFetchConcurent(ctx context.Context, srv *gmail.Service, user, query string, concurentReq int) ([]*gmail.Message, error) {
	log.Printf("searching messages from Gmail: %q", query)
	start := time.Now()

	// search
	var msgIDs []string
	err := srv.Users.Messages.List(user).Q(query).Pages(ctx, func(mr *gmail.ListMessagesResponse) error {
		for _, msg := range mr.Messages {
			msgIDs = append(msgIDs, msg.Id)
		}
		return nil
	})
	if err != nil {
		log.Printf("Unable to list messages for query:%q - %v", query, err)
		return nil, err
	}

	log.Printf("%d messages found (took %.0f sec)", len(msgIDs), time.Since(start).Seconds())
	start = time.Now()

	// parallel fetch
	bar := pb.Full.Start(len(msgIDs))
	bar.SetMaxWidth(100)
	var (
		throttle = make(chan int, concurentReq)
		wg       sync.WaitGroup
		msgs     []*gmail.Message
	)
	for i := range msgIDs {
		msgID := msgIDs[i]
		wg.Add(1)
		go func() {
			throttle <- 1
			defer func() { <-throttle; wg.Done() }()

			bar.Increment()
			msg, err := srv.Users.Messages.Get(user, msgID).Do()
			if err != nil { // TODO(bzz): retry
				log.Printf("Unable to fetch message by ID:%q", msgID)
			}

			msgs = append(msgs, msg)
		}()
	}
	wg.Wait()
	bar.Finish()

	log.Printf("%d messages fetched (took %.0f sec)", len(msgIDs), time.Since(start).Seconds())
	return msgs, nil
}

// ReadFixturesJSON reads email messages from a given JSON file.
func ReadFixturesJSON(name string) []*gmail.Message {
	var msgs []*gmail.Message
	log.Printf("reading emails from %s instead of fetching from Gmail", name)

	f, err := os.Open(name)
	if err != nil {
		log.Fatalf("Unable to read email fixtures: %v", err)
	}
	defer f.Close()

	json.NewDecoder(f).Decode(&msgs)
	return msgs
}

// ModifyMsgsDelLabel batch-deletes a label from all the given messages.
// TODO(bzz): move user to a const in this package
func ModifyMsgsDelLabel(srv *gmail.Service, user string, messages []*gmail.Message, label string) {
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
}

// FormatAsID formats human-readable lable as ID, consumable by Gmail API.
func FormatAsID(label string) string {
	// TODO(bzz): test with labels in on Gmail in Chinese/emoji
	return strings.ToLower(strings.ReplaceAll(label, " ", "-"))
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

// NormalizeAndSplit normalizes subj format and split it to type/source.
func NormalizeAndSplit(subj string) []string {
	srcType, _ := splitOnDash(subj) // handles at least EN and FR locales
	if len(srcType) != 2 {
		srcType = splitOnRuLocale(subj)
	}
	return srcType
}

type subjFormat struct{ ru, En string }

var (
	citations = subjFormat{
		": новые ссылки", "new citations",
	}
	related = subjFormat{
		"Новые статьи, связанные с работами автора ", "new related research",
	}
	search = subjFormat{
		"Новые результаты по запросу ", "new results",
	}
	articles = subjFormat{
		"Новые статьи пользователя ", "new articles",
	}
	// TODO(bzz): add this as well
	// recomended = subjFormat{
	// 	"Рекомендуемые статьи", "?????????",
	// }
)

// splitOnRuLocale normalizes subj from RU locale.
func splitOnRuLocale(s string) []string {
	var result []string

	switch {
	case strings.HasSuffix(s, citations.ru):
		result = []string{s[:strings.Index(s, citations.ru)], citations.En}
	case s == "Новые ссылки на мои статьи":
		result = []string{"me", citations.En}
	case strings.HasPrefix(s, related.ru):
		result = []string{s[len(related.ru):], related.En}
	case strings.HasPrefix(s, search.ru):
		result = []string{s[len(search.ru):], search.En}
	case strings.HasPrefix(s, articles.ru):
		result = []string{s[len(articles.ru):], articles.En}
	}

	return result
}

// splitOnDash returns str, split on unicode Dash and a separator.
func splitOnDash(str string) ([]string, string) {
	s := str
	dash := "-"
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		if unicode.In(r, unicode.Dash) {
			dash = string(r)
			break
		}
	}
	sep := fmt.Sprintf(" %s ", dash)
	return strings.Split(str, sep), sep
}

// MessageTextBody returns the text (if any) of a given message ID
func MessageTextBody(payload *gmail.MessagePart) ([]byte, error) {
	body, _, err := recursiveDecodeParts(payload, "text/html")
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

# Google Scholar alert digest
> aggregate google scholar email alerts by paper

Simplifies scientific paper discovery by aggregating all unread emails under
a Gmail label from the Google Scholar alerts, grouping papers by title and producing a report ([Markdown](https://gist.github.com/bzz/1e8445f71db03a7d57d94147279ee09f)/[HTML](https://gist.github.com/bzz/e1e3ef3e0cdabc254f4e75bfa5511bcb)).

# Workflow

 1. Search on Google Scholar for a paper of author
 2. Create an Alert (for citations, new or similar publications)
 3. Create a Gmail filter, moving all those emails under a dedicated Label
 4. Run this tool to get an aggregated report (in Markdown or HTML) of all the papers from unread emails

# Install

Using a recent version of [`git`](https://git-scm.com) and [`go`](https://golang.org)
either build a `scholar-alert-digest` binary and put it under `$GOPATH/bin` with:

```
cd "$(mktemp -d)" && go mod init scholar-alert-digest  && go get github.com/bzz/scholar-alert-digest
```

or run it directly from the clone of the sources using `go` command, as described below.

# CLI

CLI tool for Markdown/HTML report generation.

## Configure

Turn on Gmail API & download `credentials.json` following [these steps](https://developers.google.com/gmail/api/quickstart/go#step_1_turn_on_the).</br>
_That will create a new 'Quickstart' app in API console under your account and authorize it to get access to your Gmail_


To find your specific label name:

`go run main.go -labels`

To generate the report, either pass the label name though CLI:

`go run main.go -l '<your-label-name>'`

Or export it as an env var:

```shell
export SAD_LABEL='<your-label-name>'
go run main.go
```

## Run
In order to output rendered HTML instead of the default Markdown, use
```
go run main.go -html
```

To mark all emails that were aggregated in current report as read, use
```
go run main.go -mark
```

To include read emails in the separate section of the report, do
```
go run main.go -read
```

To only aggregate the email subjects do
```
go run main.go -subj | uniq -c | sort -dr
```

# Web server
Web UI that exposes basic HTML report generation to multiple concurrent users.

## Configure
It does not support the same OAuth client credentials as CLI from `credentials.json`.

It requires:
 - To create a new credentials in your API project `https://console.developers.google.com/apis/credentials?project=quickstart-<NNN>`
 - "Credentials" -> "Create credentials" -> "Web application" type
 - Add http://localhost/login/authorized value to `Authorized redirect URIs` field
 - Copy the `Client ID` and `Client secret`

Pass in the ID and the secret as env vars e.g by
```shell
export SAD_GOOGLE_ID='<client id>'
export SAD_GOOGLE_SECRET='<client secret>'
```

You do not need to pass the label name on the startup as it can be chosen at
runtime at [/labels](http://localhost:8080/labels).

## Run
The basic report generation is exposed though a web server that can be started with
```
go run ./cmd/server
```

will start a server on http://localhost:8080

Start by visiting http://localhost:8080/login to get the user OAuth access token.
Visit http://localhost:8080/labels to chose your label name.

# License

Apache License, Version 2.0. See [LICENSE](LICENSE)

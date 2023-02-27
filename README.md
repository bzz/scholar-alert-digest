# Google Scholar alert digest
> aggregate google scholar email alerts by paper

Simplifies scientific paper discovery by aggregating all unread emails under
a Gmail label from the Google Scholar alerts, grouping papers by title and producing a report ([Markdown](https://gist.github.com/bzz/1e8445f71db03a7d57d94147279ee09f)/[HTML](https://gist.github.com/bzz/e1e3ef3e0cdabc254f4e75bfa5511bcb)/[JSON](https://gist.github.com/bzz/4feeec459bcd1ec21f919eaeb163ac7a)).

* [How to use](#how-to-use)
* [Setup](#setup)
* [CLI](#cli)
* [Web server](#web-server)
* [License](#license)

# How to use

To use this tool for generating a report on new papers from Google Scholar, do the following:

 1. Search on Google Scholar for a paper of an author
 2. Create an Alert (for citations, new or similar publications)
 3. Create a Gmail filter, moving all those emails under a dedicated Label
 4. Run this tool to get an aggregated report (in Markdown or HTML) of all the papers from the unread emails

For more details, please refer to the [documentation](/docs).

# Setup

Make sure you have a recent version of [`go`](https://golang.org). Then clone this repository:

```shell
git clone github.com/bzz/scholar-alert-digest
```

## Building a binary (optional)

Alternatively, you can try to build a `scholar-alert-digest` binary and put it under `$GOPATH/bin` with:

```shell
cd "$(mktemp -d)" && go mod init scholar-alert-digest  && go get github.com/bzz/scholar-alert-digest
```

However this approach is known to yield errors and is not recommended.

## Configure google cloud

Enable "Gmail API" Google Cloud Platform (GCP) project & download `credentials.json` following [these steps](https://developers.google.com/gmail/api/quickstart/go#prerequisites).</br>
_That will guide you through creation of a new GCP project, enabling the Gmail API and geting "OAuth client ID" - [authorization credentials for a desktop application](https://developers.google.com/workspace/guides/create-credentials#oauth-client-id) that are needed in order to get access to your email messages at Gmail_

After placing credentials.json in the project directory, you need to authenticate the application. You can do this by running

```shell
go run main.go
```

An accounts.google.com link will be printed (and possibly opened in your browser). Follow the login instructions, selecting the google account you used for the previous step if you have multiple. You will get a warning that google has not verified the app; click Continue, and then Continue again. 

_Oh no! This site can't be reached!_ You'll get a "refused to connect" message. That is fine! Just go to the url bar and look for a section like this:

```
&code=4/0AWtgzh78xyaMnEMdDBL5P-tX66J3Fsb_93XvRCJzmLXDplnByMZmaXZcFjde3hJIt3D1pA
```

Copy the part following the = sign (importantly not including the trailing &scope) and paste it into the terminal. Now the app is authenticated. In the future you won't need to repeat this step

# CLI

The CLI tool is used to generate one-time Markdown/HTML reports.

To find your specific label name:

```shell
go run main.go -labels
```

To generate the report, either pass the label name though CLI:

```shell
go run main.go -l '<your-label-name>'
```

Or export it as an env var:

```shell
export SAD_LABEL='<your-label-name>'
go run main.go
```

## Run
To output rendered HTML or JSON instead of the default Markdown, use
```shell
go run main.go -html
go run main.go -json
```

To mark all emails that were aggregated in the current report as read, use
```shell
go run main.go -mark
```

To include read emails in the separate section of the report, do
```shell
go run main.go -read
```

To only aggregate the email subjects do
```
go run main.go -subj | uniq -c | sort -dr
```

There is an optional more compact report template that may be useful for a large number of papers:
```shell
go run main.go -compact
```

To include authors in the paper details snippet, use
```shell
go run main.go -authors
```

To include references to original email into the report, do:
```shell
go run main.go -refs
```

# Web Server
The Web UI exposes HTML report generation to multiple concurrent users.

## Test
It is possible to test it locally, without Gmail app configuration from below, by using emails from `./fixtures` by running:

```
go run ./cmd/server -test
```

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
The report generation is exposed through a web server that can be started with
```
go run ./cmd/server [-compact]
```

to spin up a server at http://localhost:8080

Start by visiting http://localhost:8080/login to get the user OAuth access token.
Visit http://localhost:8080/labels to chose your label name.

# License

Apache License, Version 2.0. See [LICENSE](LICENSE)

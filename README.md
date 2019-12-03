# Google Scholar alert digest
> aggregate google scholar email alerts by paper

Simplifies scientific paper discovery by aggregating all unread emails under
a Gmail label from the Google Scholar alerts, grouping papers by title and producing a report ([Markdown](https://gist.github.com/bzz/1e8445f71db03a7d57d94147279ee09f)/[HTML](https://gist.github.com/bzz/e1e3ef3e0cdabc254f4e75bfa5511bcb)).

# Workflow

 1. Search on Google Scholar for a paper of author
 2. Create an Alert (for citations, new or similar publications)
 3. Create a Gmail filter, moving all those emails under a dedicated Label
 4. Run this tool to get an aggregated report (in Markdown or HTML) of all the paper from all unread emails

# Install

Using a recent version of [`git`](https://git-scm.com) and [`go`](https://golang.org)
either build a `scholar-alert-digest` binary and put it under `$GOPATH/bin` with:

```
cd "$(mktemp -d)" && go mod init scholar-alert-digest  && go get github.com/bzz/scholar-alert-digest
```

or run from the sources by:

```
git clone https://github.com/bzz/scholar-alert-digest.git
cd scholar-alert-digest
go run main.go -h
```

# Configure

Turn on Gmail API & download `credentials.json` following [these steps](https://developers.google.com/gmail/api/quickstart/go#step_1_turn_on_the).
_It will require authorizing a 'Quickstart' app to get read-only access to your Gmail account_

# Run

To find your specific label name:

`go run main.go -labels`

To generate the report, either pass the label name though CLI:

`go run main.go -l '<your-label-name>'`

Or export it as an env var:

```shell
export SAD_LABEL='<your-label-name>'
go run main.go
```

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


# License

Apache License, Version 2.0. See [LICENSE](LICENSE)

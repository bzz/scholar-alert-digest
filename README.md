# Google Scholar alert digest
> aggregate google scholar email alerts by paper

Simplifies scientific paper discovery by aggregating all unread emails under
a Gmail label from the Google Scholar alerts, grouping papers by title and producing a report in Markdown.

# Workflow

 1. Search on Google Scholar for a paper of author
 2. Create an Alert (for citations, new or similar publications)
 3. Create a GMail filter, moving all those emails under a dedicated Label
 4. Run this tool to get an aggregated Markdown report of all the paper from all unread emails

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

# License

Apache License, Version 2.0. See [LICENSE](LICENSE)
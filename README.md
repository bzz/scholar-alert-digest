# Google Scholar alert digest
> aggregate google scholar email alerts by paper

Simplifies scientific paper discovery by aggregating all unread emails under
Gmail label from Google Scholar, grouping papers by title and formating in Markdown.

# Workflow

 1. Search on Google Scholar for a paper of author
 2. Create an Alert (for citations, new or similar publications)
 3. Create a GMail filter, moving all such emails under a dedicated Label
 4. Run this tool to get a Markdown digest of all the paper from unread emails

# Run

`go run main.go [-l <your-label>]`

# License

Apache License, Version 2.0. See [LICENSE](LICENSE)
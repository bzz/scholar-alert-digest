# High-level overview of the application

The application does the following steps:

 - check if OAuth _Token_ is present in the  session, if not - redirect to Google web OAuth
 - check if Gmail _Label_ name is present in the session, if not - fetch all labels and choose one on `/labels`
 - fetch un-read emails under a given _Label_ using _Token_ though GMail API (using a [search query](https://github.com/bzz/scholar-alert-digest/blob/c4600bfa4faf8cfc4347e31dc1489a26b0a95222/cmd/server/server.go#L132)).
 - (optional) fetch un-read emails the same way (if enabled by `-read`, only supported by CLI now)
 - extract paper mentions from each read email, sort and aggregate by paper title, count frequency of the duplicates
 - (optional) fetch&extract the read emails (enabled by `-read`)
 - render read papers using the [tempates](https://github.com/bzz/scholar-alert-digest/blob/c4600bfa4faf8cfc4347e31dc1489a26b0a95222/templates/templates.go#L20) in one of the supported formats (JSONL, Markdown, HTML)
 - (optional) render read emails (in separate "Archive" section, enabled by `-read`)
 - (optional) mark all un-read emails that were used to produce this report as read on gmail (enabled by `-mark`, only supported by CLI now)


## Data model

The application data model consits of:

(one) email **Message**
 * Subject (plain-text, user local language, follows one of [the number of templates](https://github.com/bzz/scholar-alert-digest/blob/c4600bfa4faf8cfc4347e31dc1489a26b0a95222/gmailutils/gmail.go#L253))
 * Body (MIME encoded parts)


that contains, or produces

(many) **Paper**s
 * Title, URL, Abstract
 * Author (only displayed if enabled by `-author`, on by default on server)
 * Refs[] (`[{ID, Title}, ...]` all emails that are "origins of the citation" or "sources, refering to" this paper)
 * Freq (citation frequency: a total number of Messages reffering to this paper)


The reference information in a paper is utilized to produce links to the original email messages on Gmail e.g. for user to mark them as "read" in manually, if no `-mark` is provided (wich is only supported by CLI)


Statistics in the header reports the following:
 * datetime when the report was generated
 * total number of unread email *Messages* that were used to generate it
 * total number of *Papers* (by unique paper titles) cited/references in these unread emails

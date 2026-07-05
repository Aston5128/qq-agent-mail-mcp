package agently

import "strconv"

type EmptyInput struct{}

func (EmptyInput) Args() []string {
	return []string{"+me"}
}

type MessageListInput struct {
	Limit          int    `json:"limit,omitempty" jsonschema:"messages per page, max 50"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"pagination cursor"`
	Dir            string `json:"dir,omitempty" jsonschema:"folder: inbox, sent, trash, spam"`
	After          string `json:"after,omitempty" jsonschema:"only messages after this ISO 8601 time"`
	Before         string `json:"before,omitempty" jsonschema:"only messages before this ISO 8601 time"`
	IsUnread       bool   `json:"is_unread,omitempty" jsonschema:"only show unread messages"`
	HasAttachments bool   `json:"has_attachments,omitempty" jsonschema:"only show messages with attachments"`
}

func (in MessageListInput) Args() []string {
	args := []string{"message", "+list"}
	addCommonListFlags(&args, in.Limit, in.Cursor, in.Dir, in.After, in.Before, in.IsUnread, in.HasAttachments)
	return args
}

type MessageReadInput struct {
	ID string `json:"id" jsonschema:"message_id with msg_ prefix"`
}

func (in MessageReadInput) Args() []string {
	return []string{"message", "+read", "--id", in.ID}
}

type MessageSearchInput struct {
	Query          string `json:"q,omitempty" jsonschema:"search keyword or phrase"`
	SearchIn       string `json:"search_in,omitempty" jsonschema:"SEARCH_IN_ALL, SEARCH_IN_SUBJECT, or SEARCH_IN_CONTENT"`
	From           string `json:"from,omitempty" jsonschema:"sender email filter"`
	To             string `json:"to,omitempty" jsonschema:"recipient email filter"`
	Limit          int    `json:"limit,omitempty" jsonschema:"results per page, max 50"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"pagination cursor"`
	Dir            string `json:"dir,omitempty" jsonschema:"folder: inbox, sent, trash, spam"`
	After          string `json:"after,omitempty" jsonschema:"only messages after this ISO 8601 time"`
	Before         string `json:"before,omitempty" jsonschema:"only messages before this ISO 8601 time"`
	IsUnread       bool   `json:"is_unread,omitempty" jsonschema:"only show unread messages"`
	HasAttachments bool   `json:"has_attachments,omitempty" jsonschema:"only show messages with attachments"`
}

func (in MessageSearchInput) Args() []string {
	args := []string{"message", "+search"}
	addString(&args, "--q", in.Query)
	addString(&args, "--search-in", in.SearchIn)
	addString(&args, "--from", in.From)
	addString(&args, "--to", in.To)
	addCommonListFlags(&args, in.Limit, in.Cursor, in.Dir, in.After, in.Before, in.IsUnread, in.HasAttachments)
	return args
}

type MessageSendInput struct {
	To                []string `json:"to" jsonschema:"recipient email addresses"`
	Cc                []string `json:"cc,omitempty" jsonschema:"CC recipient email addresses"`
	Bcc               []string `json:"bcc,omitempty" jsonschema:"BCC recipient email addresses"`
	Subject           string   `json:"subject" jsonschema:"email subject"`
	Body              string   `json:"body,omitempty" jsonschema:"email body"`
	BodyFile          string   `json:"body_file,omitempty" jsonschema:"relative UTF-8 body file path"`
	BodyFormat        string   `json:"body_format,omitempty" jsonschema:"body format, e.g. plain"`
	Attachment        []string `json:"attachment,omitempty" jsonschema:"relative attachment file paths"`
	ConfirmationToken string   `json:"confirmation_token,omitempty" jsonschema:"token returned by first send call"`
}

func (in MessageSendInput) Args() []string {
	args := []string{"message", "+send"}
	addRepeated(&args, "--to", in.To)
	addRepeated(&args, "--cc", in.Cc)
	addRepeated(&args, "--bcc", in.Bcc)
	addString(&args, "--subject", in.Subject)
	addString(&args, "--body", in.Body)
	addString(&args, "--body-file", in.BodyFile)
	addString(&args, "--body-format", in.BodyFormat)
	addRepeated(&args, "--attachment", in.Attachment)
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	return args
}

type MessageReplyInput struct {
	ID                string   `json:"id" jsonschema:"message_id to reply to"`
	Body              string   `json:"body,omitempty" jsonschema:"reply body"`
	BodyFile          string   `json:"body_file,omitempty" jsonschema:"relative UTF-8 body file path"`
	BodyFormat        string   `json:"body_format,omitempty" jsonschema:"body format, e.g. plain"`
	Cc                []string `json:"cc,omitempty" jsonschema:"additional CC recipients"`
	Bcc               []string `json:"bcc,omitempty" jsonschema:"additional BCC recipients"`
	Attachment        []string `json:"attachment,omitempty" jsonschema:"relative attachment file paths"`
	ConfirmationToken string   `json:"confirmation_token,omitempty" jsonschema:"token returned by first reply call"`
	ReplyAll          bool     `json:"reply_all,omitempty" jsonschema:"reply to all recipients"`
}

func (in MessageReplyInput) Args() []string {
	args := []string{"message", "+reply", "--id", in.ID}
	addString(&args, "--body", in.Body)
	addString(&args, "--body-file", in.BodyFile)
	addString(&args, "--body-format", in.BodyFormat)
	addRepeated(&args, "--cc", in.Cc)
	addRepeated(&args, "--bcc", in.Bcc)
	addRepeated(&args, "--attachment", in.Attachment)
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	addBool(&args, "--reply-all", in.ReplyAll)
	return args
}

type MessageForwardInput struct {
	ID                 string   `json:"id" jsonschema:"message_id to forward"`
	To                 []string `json:"to" jsonschema:"forward recipient email addresses"`
	Cc                 []string `json:"cc,omitempty" jsonschema:"CC recipient email addresses"`
	Bcc                []string `json:"bcc,omitempty" jsonschema:"BCC recipient email addresses"`
	Body               string   `json:"body,omitempty" jsonschema:"forwarding note"`
	BodyFile           string   `json:"body_file,omitempty" jsonschema:"relative UTF-8 body file path"`
	BodyFormat         string   `json:"body_format,omitempty" jsonschema:"body format, e.g. plain"`
	Attachment         []string `json:"attachment,omitempty" jsonschema:"additional relative attachment paths"`
	ConfirmationToken  string   `json:"confirmation_token,omitempty" jsonschema:"token returned by first forward call"`
	IncludeAttachments bool     `json:"include_attachments,omitempty" jsonschema:"include original message attachments"`
}

func (in MessageForwardInput) Args() []string {
	args := []string{"message", "+forward", "--id", in.ID}
	addRepeated(&args, "--to", in.To)
	addRepeated(&args, "--cc", in.Cc)
	addRepeated(&args, "--bcc", in.Bcc)
	addString(&args, "--body", in.Body)
	addString(&args, "--body-file", in.BodyFile)
	addString(&args, "--body-format", in.BodyFormat)
	addRepeated(&args, "--attachment", in.Attachment)
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	addBool(&args, "--include-attachments", in.IncludeAttachments)
	return args
}

type MessageTrashInput struct {
	ID                string `json:"id" jsonschema:"message_id to move to trash"`
	ConfirmationToken string `json:"confirmation_token,omitempty" jsonschema:"token returned by first trash call"`
}

func (in MessageTrashInput) Args() []string {
	args := []string{"message", "+trash", "--id", in.ID}
	addString(&args, "--confirmation-token", in.ConfirmationToken)
	return args
}

type AttachmentUploadInput struct {
	File string `json:"file" jsonschema:"relative local file path to upload"`
}

func (in AttachmentUploadInput) Args() []string {
	return []string{"attachment", "+upload", "--file", in.File}
}

type AttachmentDownloadInput struct {
	MessageID    string `json:"message_id" jsonschema:"message_id with msg_ prefix"`
	AttachmentID string `json:"attachment_id" jsonschema:"attachment_id with att_ prefix"`
	Output       string `json:"output,omitempty" jsonschema:"relative output directory"`
}

func (in AttachmentDownloadInput) Args() []string {
	args := []string{"attachment", "+download", "--msg", in.MessageID, "--att", in.AttachmentID}
	addString(&args, "--output", in.Output)
	return args
}

func addCommonListFlags(args *[]string, limit int, cursor, dir, after, before string, isUnread, hasAttachments bool) {
	if limit > 0 {
		*args = append(*args, "--limit", strconv.Itoa(limit))
	}
	addString(args, "--cursor", cursor)
	addString(args, "--dir", dir)
	addString(args, "--after", after)
	addString(args, "--before", before)
	addBool(args, "--is-unread", isUnread)
	addBool(args, "--has-attachments", hasAttachments)
}

func addString(args *[]string, flag string, value string) {
	if value != "" {
		*args = append(*args, flag, value)
	}
}

func addRepeated(args *[]string, flag string, values []string) {
	for _, value := range values {
		if value != "" {
			*args = append(*args, flag, value)
		}
	}
}

func addBool(args *[]string, flag string, value bool) {
	if value {
		*args = append(*args, flag)
	}
}

package agently

import (
	"reflect"
	"testing"
)

func TestMessageListArgs(t *testing.T) {
	got := MessageListInput{Limit: 10, Dir: "inbox", HasAttachments: true}.Args()
	want := []string{"message", "+list", "--limit", "10", "--dir", "inbox", "--has-attachments"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestMessageSendArgs(t *testing.T) {
	got := MessageSendInput{
		To:                []string{"a@example.com", "b@example.com"},
		Subject:           "Hi",
		Body:              "Hello",
		ConfirmationToken: "tok_123",
	}.Args()
	want := []string{
		"message", "+send",
		"--to", "a@example.com",
		"--to", "b@example.com",
		"--subject", "Hi",
		"--body", "Hello",
		"--confirmation-token", "tok_123",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestMessageTrashArgs(t *testing.T) {
	got := MessageTrashInput{ID: "msg_1"}.Args()
	want := []string{"message", "+trash", "--id", "msg_1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestMessageTrashArgsWithConfirmationToken(t *testing.T) {
	got := MessageTrashInput{ID: "msg_1", ConfirmationToken: "tok_123"}.Args()
	want := []string{"message", "+trash", "--id", "msg_1", "--confirmation-token", "tok_123"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestAttachmentDownloadArgs(t *testing.T) {
	got := AttachmentDownloadInput{
		MessageID:    "msg_1",
		AttachmentID: "att_1",
		Output:       "downloads",
	}.Args()
	want := []string{"attachment", "+download", "--msg", "msg_1", "--att", "att_1", "--output", "downloads"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

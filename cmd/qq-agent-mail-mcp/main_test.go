package main

import "testing"

func TestGetenvReturnsFallback(t *testing.T) {
	t.Setenv("QQ_AGENT_MAIL_MCP_TEST_EMPTY", "")
	if got := getenv("QQ_AGENT_MAIL_MCP_TEST_EMPTY", "fallback"); got != "fallback" {
		t.Fatalf("getenv returned %q, want fallback", got)
	}
}

func TestGetenvReturnsValue(t *testing.T) {
	t.Setenv("QQ_AGENT_MAIL_MCP_TEST_VALUE", "custom")
	if got := getenv("QQ_AGENT_MAIL_MCP_TEST_VALUE", "fallback"); got != "custom" {
		t.Fatalf("getenv returned %q, want custom", got)
	}
}

package handlers

import (
	"strings"
	"testing"
)

func TestNormalizeChatMessageUsesDatabaseCharacterLimit(t *testing.T) {
	message, valid := normalizeChatMessage("  안녕하세요  ")
	if !valid || message != "안녕하세요" {
		t.Fatalf("normalized message = %q, valid = %v", message, valid)
	}
	if _, valid := normalizeChatMessage(strings.Repeat("가", 2000)); !valid {
		t.Fatal("2000 Unicode characters were rejected")
	}
	if _, valid := normalizeChatMessage(strings.Repeat("가", 2001)); valid {
		t.Fatal("2001 Unicode characters were accepted")
	}
	if _, valid := normalizeChatMessage(" \n\t "); valid {
		t.Fatal("whitespace-only message was accepted")
	}
}

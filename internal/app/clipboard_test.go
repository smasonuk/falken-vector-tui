package app

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type clipboardCall struct {
	name  string
	args  []string
	input string
}

func TestCopyToClipboardDarwinUsesPbcopyAndExactText(t *testing.T) {
	var calls []clipboardCall
	input := " first line\nsecond line \n"
	err := copyToClipboardWith("darwin", input, func(name string, args []string, text string) error {
		calls = append(calls, clipboardCall{name: name, args: args, input: text})
		return nil
	})
	if err != nil {
		t.Fatalf("copyToClipboardWith: %v", err)
	}
	want := []clipboardCall{{name: "pbcopy", input: input}}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %+v, want %+v", calls, want)
	}
}

func TestCopyToClipboardLinuxFallsBack(t *testing.T) {
	var calls []clipboardCall
	err := copyToClipboardWith("linux", "text", func(name string, args []string, text string) error {
		calls = append(calls, clipboardCall{name: name, args: append([]string(nil), args...), input: text})
		if name == "xclip" {
			return nil
		}
		return errors.New("not found")
	})
	if err != nil {
		t.Fatalf("copyToClipboardWith: %v", err)
	}
	want := []clipboardCall{
		{name: "wl-copy", input: "text"},
		{name: "xclip", args: []string{"-selection", "clipboard"}, input: "text"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %+v, want %+v", calls, want)
	}
}

func TestCopyToClipboardWindowsUsesClip(t *testing.T) {
	var calls []clipboardCall
	err := copyToClipboardWith("windows", "text", func(name string, args []string, text string) error {
		calls = append(calls, clipboardCall{name: name, args: args, input: text})
		return nil
	})
	if err != nil {
		t.Fatalf("copyToClipboardWith: %v", err)
	}
	want := []clipboardCall{{name: "clip", input: "text"}}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %+v, want %+v", calls, want)
	}
}

func TestCopyToClipboardEmptyTextIsNoOp(t *testing.T) {
	called := false
	err := copyToClipboardWith("darwin", "", func(name string, args []string, text string) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("copyToClipboardWith: %v", err)
	}
	if called {
		t.Fatal("runner called for empty text")
	}
}

func TestCopyToClipboardReportsMissingCommand(t *testing.T) {
	err := copyToClipboardWith("darwin", "text", func(name string, args []string, text string) error {
		return errors.New("not found")
	})
	if err == nil {
		t.Fatal("copyToClipboardWith err = nil, want error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "clipboard command not found", "pbcopy") {
		t.Fatalf("err = %q", got)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}

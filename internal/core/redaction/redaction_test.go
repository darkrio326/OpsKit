package redaction

import (
	"slices"
	"strings"
	"testing"
)

func TestRedactText(t *testing.T) {
	in := `password=abc token:xyz "secret":"v1" --token=aaa --password bbb`
	out := RedactText(in)
	if out == in {
		t.Fatalf("expected redaction to change text")
	}
	for _, plain := range []string{"abc", "xyz", "v1", "aaa", "bbb"} {
		if strings.Contains(out, plain) {
			t.Fatalf("expected plaintext %q to be redacted", plain)
		}
	}
	if out == "" {
		t.Fatalf("unexpected empty output")
	}
}

func TestRedactArgs(t *testing.T) {
	args := []string{"--password", "abc", "token=def", "--name", "ok", "--db.secret", "ghi"}
	got := RedactArgs(args)
	want := []string{"--password", Mask, "token=" + Mask, "--name", "ok", "--db.secret", Mask}
	if len(got) != len(want) {
		t.Fatalf("unexpected arg len: %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg[%d] expected %q got %q", i, want[i], got[i])
		}
	}
}

func TestDefaultKeys(t *testing.T) {
	keys := DefaultKeys()
	for _, expect := range []string{"password", "token", "secret"} {
		if !slices.Contains(keys, expect) {
			t.Fatalf("missing default key: %s", expect)
		}
	}
}

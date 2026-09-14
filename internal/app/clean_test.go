package app

import (
	"strings"
	"testing"
	"unicode"

	"github.com/MrZloHex/monolink"
	"github.com/MrZloHex/monolink/marshal"
)

// An answer to a Request never passes through plain: it comes back to
// whoever asked. Everything another node says must still be clean before it
// is stored, or a terminal acts on an escape sequence among it — redraws
// itself, retitles the window, writes the clipboard.

const esc = "\x1b]0;taken\x07" // set the window title, as a terminal reads it

// noControls is what cleaning guarantees: the characters a terminal acts on
// are gone. What is left of a sequence — "]0;taken" — is plain text now, and
// is left alone rather than guessed at.
func noControls(t *testing.T, what, s string) {
	t.Helper()
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) && r != '\u200c' && r != '\u200d' {
			t.Fatalf("%s kept %q: %q", what, r, s)
		}
	}
}

func TestAnErrDetailIsCleanedBeforeItIsShown(t *testing.T) {
	err := &monolink.ReplyError{Code: monolink.CodeState, Detail: "no such thing" + esc,
		Reply: monolink.Message{From: "GOVERNOR"}}
	got := describe(err)
	noControls(t, "describe", got)
	if !strings.Contains(got, "no such thing") {
		t.Fatalf("describe lost the detail: %q", got)
	}
}

func TestEveryErrCodeIsCleaned(t *testing.T) {
	for _, code := range []string{monolink.CodeDenied, monolink.CodeBusy, monolink.CodeNAC,
		monolink.CodeState, monolink.CodeArg, monolink.CodeInternal} {
		err := &monolink.ReplyError{Code: code, Detail: "why" + esc, Reply: monolink.Message{From: "MARSHAL"}}
		noControls(t, "ERR "+code, describe(err))
	}
}

func TestAMessageIsCleanedAndItsPeopleChecked(t *testing.T) {
	header := monolink.Record("7", "dasha", "mzh", "2026-09-13T12:00:00Z", "OFF")
	msgs, err := parseMsgs([]string{header, "hello" + esc})
	if err != nil {
		t.Fatal(err)
	}
	noControls(t, "a message", msgs[0].text)
	if !strings.HasPrefix(msgs[0].text, "hello") {
		t.Fatalf("a message lost its text: %q", msgs[0].text)
	}

	// A name is checked, not cleaned: taking the escape out of "da<esc>sha"
	// would file a stranger's message under dasha.
	bad := monolink.Record("8", "da\x1bsha", "mzh", "2026-09-13T12:00:00Z", "OFF")
	if _, err := parseMsgs([]string{bad, "hello"}); err == nil {
		t.Fatal("parseMsgs took a sender whose name nobody can have")
	}
}

func TestAConversationWithAnImpossibleNameIsRefused(t *testing.T) {
	good := monolink.Record("dasha", "2", "9", "2026-09-13T12:00:00Z")
	if _, err := parseChats([]string{good}); err != nil {
		t.Fatal(err)
	}
	bad := monolink.Record("da\x1bsha", "2", "9", "2026-09-13T12:00:00Z")
	if _, err := parseChats([]string{bad}); err == nil {
		t.Fatal("parseChats took a peer whose name nobody can have")
	}
}

func TestPeopleAreKeptOnlyIfTheirNamesAreOnes(t *testing.T) {
	people, err := parsePeople(monolink.Record("mzh", "da\x1bsha", "lena"))
	if err != nil {
		t.Fatal(err)
	}
	if len(people) != 2 || people[0] != "mzh" || people[1] != "lena" {
		t.Fatalf("parsePeople kept the wrong people: %q", people)
	}
}

func TestGrantsAndKeysAndSessionsAreCleaned(t *testing.T) {
	noControls(t, "a grant", cleanAll([]string{"VERTEX.*" + esc})[0])
	keys := cleanKeys([]marshal.KeyInfo{{Ref: "ab12", Kind: "webauthn", Label: "phone" + esc}})
	noControls(t, "a key label", keys[0].Label)
	ss := cleanSessions([]marshal.SessionInfo{{User: "mzh" + esc, Panel: "MONOWEB" + esc}})
	noControls(t, "a session's person", ss[0].User)
	noControls(t, "a session's panel", ss[0].Panel)
}

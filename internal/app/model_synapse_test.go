package app

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrZloHex/monolink"
	"github.com/MrZloHex/monolink/marshal"
	"monoview/internal/types"
)

// A long message goes as several, each fitting one field once escaped, and
// nothing is lost at the cuts.
func TestSplitMessageKeepsEveryPartInOneField(t *testing.T) {
	long := strings.Repeat("Купи молоко: 2 л, хлеб — 100% ", 30)
	for _, text := range []string{long, strings.Repeat("ж", 300), strings.Repeat("::", 200)} {
		parts := splitMessage(text)
		if len(parts) < 2 {
			t.Fatalf("%d bytes went as %d message", len(text), len(parts))
		}
		for _, p := range parts {
			if n := len(monolink.Escape(p)); n > monolink.MaxField || p == "" || p != strings.TrimSpace(p) {
				t.Fatalf("part of %d escaped bytes: %q", n, p)
			}
		}
		if strings.Join(strings.Fields(strings.Join(parts, " ")), "") != strings.Join(strings.Fields(text), "") {
			t.Fatalf("text lost in the cut: %q", parts)
		}
	}
	if got := splitMessage("  hi  "); len(got) != 1 || got[0] != "hi" {
		t.Fatalf("short message split as %q", got)
	}
	if got := splitMessage("   "); len(got) != 0 {
		t.Fatalf("blank message split as %q", got)
	}
}

func TestParseMsgsReadsSynapsesPages(t *testing.T) {
	args := []string{
		monolink.Record("7", "mzh", "dasha", "2026-09-11T18:30:00+03:00", "ON"), "at 18:30 | ok",
		monolink.Record("9", "dasha", "mzh", "2026-09-11T18:31:00+03:00", "OFF"), "да",
	}
	got, err := parseMsgs(args)
	if err != nil || len(got) != 2 {
		t.Fatalf("%+v, %v", got, err)
	}
	if m := got[0]; m.id != 7 || m.from != "mzh" || m.to != "dasha" || !m.read || m.text != "at 18:30 | ok" || m.at.Minute() != 30 {
		t.Fatalf("first message %+v", m)
	}
	if got[1].read || got[1].text != "да" {
		t.Fatalf("second message %+v", got[1])
	}
	if _, err := parseMsgs(args[:3]); err == nil {
		t.Fatal("an odd number of fields was accepted")
	}
}

func TestPagesFoldIntoTheConversation(t *testing.T) {
	page := func(ids ...uint64) []synapseMsg {
		var out []synapseMsg
		for _, id := range ids {
			out = append(out, synapseMsg{id: id})
		}
		return out
	}
	idsOf := func(msgs []synapseMsg) (out []uint64) {
		for _, m := range msgs {
			out = append(out, m.id)
		}
		return
	}
	full := page(13, 14, 15, 16, 17, 18, 19, 20)

	have, older := mergePage(nil, full, "", "", false)
	if len(have) != 8 || !older {
		t.Fatalf("the newest page: %v, older %v", idsOf(have), older)
	}
	have, older = mergePage(have, page(21, 22), "20", "", older)
	if got := idsOf(have); len(got) != 10 || got[9] != 22 || !older {
		t.Fatalf("after a newer page: %v, older %v", got, older)
	}
	have, older = mergePage(have, page(11, 12), "", "13", older)
	if got := idsOf(have); got[0] != 11 || len(got) != 12 || older {
		t.Fatalf("after the first page: %v, older %v", got, older)
	}
	// A full newest page past what is shown leaves a gap: start again from it.
	have, _ = mergePage(have, page(40, 41, 42, 43, 44, 45, 46, 47), "", "", older)
	if got := idsOf(have); got[0] != 40 || len(got) != 8 {
		t.Fatalf("across a gap: %v", got)
	}
}

func signedInAt(sheet types.Sheet) Model {
	m := NewModel()
	m.Width, m.Height = 140, 40
	m.adopt(marshal.Session{Token: "t", User: "mzh", Expires: time.Now().Add(time.Hour)}, []string{"*"})
	m.ActiveSheet = sheet
	return m
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case " ":
			msg = tea.KeyMsg{Type: tea.KeySpace}
		default:
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

// While a message is being written every key is text: a 5 does not switch
// sheets and a q does not quit.
func TestWritingTakesEveryKey(t *testing.T) {
	m := signedInAt(types.SheetSynapse)
	m.Synapse.people = []string{"dasha", "mzh"}
	m = press(m, "i")
	if !m.Synapse.writing || m.Synapse.peer != "dasha" {
		t.Fatalf("[i] did not open a message to dasha: %+v", m.Synapse)
	}
	m = press(m, "5", " ", "q", "!")
	if m.ActiveSheet != types.SheetSynapse || m.Synapse.draft != "5 q!" {
		t.Fatalf("sheet %v, draft %q", m.ActiveSheet, m.Synapse.draft)
	}
	m = press(m, "esc")
	if m.Synapse.writing || m.Synapse.draft != "5 q!" {
		t.Fatalf("Esc: writing %v, draft %q", m.Synapse.writing, m.Synapse.draft)
	}
	m = press(m, "4")
	if m.ActiveSheet != types.SheetSystem {
		t.Fatal("keys went on being text after Esc")
	}
}

// The tab shows what SYNAPSE says is unread for the person signed in, and
// nobody else's count.
func TestUnreadFollowsSynapse(t *testing.T) {
	m := signedInAt(types.SheetCalendar)
	pub := func(noun, v string) {
		m.handleSynapsePub(monolink.Message{Version: monolink.V2, From: "SYNAPSE", To: "ALL", Verb: "PUB", Noun: noun, Args: []string{v}})
	}
	pub("UNREAD.mzh", "3")
	pub("UNREAD.dasha", "9")
	if m.Synapse.unread != 3 || !strings.Contains(m.renderTabs(), "[6] SYNAPSE (3)") {
		t.Fatalf("unread %d; tabs %q", m.Synapse.unread, m.renderTabs())
	}
	pub("UNREAD.mzh", "0")
	if strings.Contains(m.renderTabs(), "SYNAPSE (") {
		t.Fatal("a count stayed on the tab after everything was read")
	}
}

func TestSynapseNeedsSomeoneSignedIn(t *testing.T) {
	m := NewModel()
	m.Width, m.Height = 140, 40
	m.ActiveSheet = types.SheetSynapse
	if !strings.Contains(m.renderSynapse(), "Sign in") {
		t.Fatal("the sheet does not say it needs someone signed in")
	}
	if cmd := m.synapseRefresh(); cmd != nil {
		t.Fatal("asked SYNAPSE with nobody signed in")
	}
}

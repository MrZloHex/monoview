package app

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrZloHex/monolink"
	"github.com/MrZloHex/monolink/marshal"
	"monoview/internal/types"
)

// The SYNAPSE sheet: messages between the people of the bubble (SPEC §29).
// SYNAPSE files a message under the person in the sender's address, so the
// sheet works only with someone signed in, and writes as MONOVIEW.<person>.

const (
	synapseNode = "SYNAPSE"
	synapsePage = monolink.MaxArgs / 2 // messages in one of SYNAPSE's pages
)

type synapseMsg struct {
	id       uint64
	from, to string
	at       time.Time
	read     bool // by to
	text     string
}

type synapseChat struct {
	unread int
	last   uint64
	at     time.Time
}

// synapseState is everything the sheet shows.
type synapseState struct {
	people   []string               // who can be written to: SYNAPSE's PEOPLE
	chats    map[string]synapseChat // by peer
	unread   int                    // UNREAD.<me>
	selected int                    // index into synapsePeers()
	peer     string                 // the open conversation; "" for none
	msgs     []synapseMsg           // of the open conversation, oldest first
	older    bool                   // earlier messages may exist
	writing  bool                   // typing to peer; takes every key
	draft    string
	busy     bool // a message is on its way
	status   string
}

// Answers from SYNAPSE, delivered as tea messages: every call runs in a
// command, never in Update.
type (
	synapseListMsg struct {
		people []string
		chats  map[string]synapseChat
		err    error
	}
	synapseConvMsg struct {
		peer          string
		after, before string // which page: both empty is the newest
		msgs          []synapseMsg
		err           error
	}
	synapseSentMsg struct {
		peer string
		err  error
	}
)

func synapseAsk(ctx context.Context, hub *monolink.Client, verb, noun string, args ...string) (monolink.Message, error) {
	return hub.RequestDialect(ctx, monolink.V2, synapseNode, verb, noun, args...)
}

// synapseRefresh asks whom this person can write to, and how each of their
// conversations stands.
func (m *Model) synapseRefresh() tea.Cmd {
	if !m.signedIn() || !m.permitted(synapseNode, monolink.VerbGet, "CHATS") {
		return nil
	}
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		r, err := synapseAsk(ctx, hub, monolink.VerbGet, "PEOPLE")
		if err != nil {
			return synapseListMsg{err: err}
		}
		people, err := parsePeople(r.Arg(0))
		if err != nil {
			return synapseListMsg{err: err}
		}
		chats := map[string]synapseChat{}
		var before []string
		for range 8 { // as many pages as there are, within reason
			c, err := synapseAsk(ctx, hub, monolink.VerbGet, "CHATS", before...)
			if err != nil {
				return synapseListMsg{err: err}
			}
			page, err := parseChats(c.Args)
			if err != nil {
				return synapseListMsg{err: err}
			}
			for peer, chat := range page {
				chats[peer] = chat
			}
			next := chatsBefore(len(c.Args), page)
			if next == "" {
				break
			}
			before = []string{next}
		}
		return synapseListMsg{people: people, chats: chats}
	})
}

// chatsBefore is where the CHATS page after this one starts — its oldest
// conversation's last id — or "" if this page, of n fields, was the last: a
// full page is a frame's worth.
func chatsBefore(n int, page map[string]synapseChat) string {
	if n < monolink.MaxArgs {
		return ""
	}
	var oldest uint64
	for _, c := range page {
		if c.last > 0 && (oldest == 0 || c.last < oldest) {
			oldest = c.last
		}
	}
	if oldest == 0 {
		return ""
	}
	return strconv.FormatUint(oldest, 10)
}

// synapseOpen shows the conversation with peer, from its newest page.
func (m *Model) synapseOpen(peer string) tea.Cmd {
	p := &m.Synapse
	if p.peer != peer {
		p.peer, p.msgs, p.older, p.draft = peer, nil, false, ""
	}
	return m.synapseLoad(peer, "", "")
}

// synapseLoad fetches a page of the conversation with peer — the newest, or
// the one after or before an id — and marks what peer wrote as read.
func (m *Model) synapseLoad(peer, after, before string) tea.Cmd {
	if !m.permitted(synapseNode, monolink.VerbGet, "MSGS") {
		return nil
	}
	markRead := m.permitted(synapseNode, monolink.VerbDo, "READ.MSG")
	args := []string{peer}
	switch {
	case before != "":
		args = append(args, after, before)
	case after != "":
		args = append(args, after)
	}
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		out := synapseConvMsg{peer: peer, after: after, before: before}
		r, err := synapseAsk(ctx, hub, monolink.VerbGet, "MSGS", args...)
		if err != nil {
			out.err = err
			return out
		}
		out.msgs, out.err = parseMsgs(r.Args)
		if out.err == nil && markRead {
			if id := lastUnreadFrom(out.msgs, peer); id != 0 {
				// Reading one reads everything before it; UNREAD.<me> follows.
				synapseAsk(ctx, hub, monolink.VerbDo, "READ.MSG", strconv.FormatUint(id, 10))
			}
		}
		return out
	})
}

// synapseSend sends the draft to the open conversation: as several messages
// if it is longer than one field, in order, one after another.
func (m *Model) synapseSend() tea.Cmd {
	p := &m.Synapse
	parts := splitMessage(p.draft)
	if len(parts) == 0 || p.peer == "" {
		return nil
	}
	if !m.permitted(synapseNode, monolink.VerbDo, "SEND.MSG") {
		p.status = "not permitted: SYNAPSE.DO.SEND.MSG"
		return nil
	}
	peer := p.peer
	p.busy = true
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		for _, part := range parts {
			if _, err := synapseAsk(ctx, hub, monolink.VerbDo, "SEND.MSG", peer, part); err != nil {
				return synapseSentMsg{peer: peer, err: err}
			}
		}
		return synapseSentMsg{peer: peer}
	})
}

func (m *Model) handleSynapseMsg(msg tea.Msg) tea.Cmd {
	p := &m.Synapse
	switch msg := msg.(type) {
	case synapseListMsg:
		if msg.err != nil {
			p.status = synapseDescribe(msg.err)
			return nil
		}
		p.people, p.chats, p.status = msg.people, msg.chats, ""
		p.unread = 0
		for _, c := range msg.chats {
			p.unread += c.unread
		}
		if n := len(m.synapsePeers()); p.selected >= n {
			p.selected = max(n-1, 0)
		}

	case synapseConvMsg:
		if msg.peer != p.peer {
			return nil // the conversation was left meanwhile
		}
		if msg.err != nil {
			p.status = synapseDescribe(msg.err)
			return nil
		}
		p.msgs, p.older = mergePage(p.msgs, msg.msgs, msg.after, msg.before, p.older)

	case synapseSentMsg:
		p.busy = false
		if msg.err != nil {
			p.status = "not sent: " + synapseDescribe(msg.err)
			return nil
		}
		p.draft, p.status = "", ""
		if msg.peer == p.peer {
			return m.synapseLoad(p.peer, lastID(p.msgs), "")
		}
	}
	return nil
}

// handleSynapsePub follows what SYNAPSE publishes: UNREAD.<me> moving means
// something arrived for this person or was read by them; UNREAD.<peer>
// moving means the open conversation's receipts may have changed.
func (m *Model) handleSynapsePub(msg monolink.Message) tea.Cmd {
	if msg.Verb != monolink.VerbPub || !m.signedIn() {
		return nil
	}
	if a, err := monolink.ParseAddress(msg.From); err != nil || a.Node != synapseNode {
		return nil
	}
	p := &m.Synapse
	switch msg.Noun {
	case "PEOPLE":
		if people, err := parsePeople(msg.Arg(0)); err == nil {
			p.people = people
		}
	case "UNREAD." + m.Session.User:
		if n, err := strconv.Atoi(msg.Arg(0)); err == nil {
			p.unread = n
		}
		cmds := []tea.Cmd{m.synapseRefresh()}
		if p.peer != "" {
			cmds = append(cmds, m.synapseLoad(p.peer, lastID(p.msgs), ""))
		}
		return tea.Batch(cmds...)
	case "UNREAD." + p.peer:
		if p.peer != "" {
			return m.synapseLoad(p.peer, "", "")
		}
	}
	return nil
}

// ─── keys ────────────────────────────────────────────────────────────

// handleSynapseKeys serves the SYNAPSE sheet and, while a message is being
// written, takes every key — so a message may contain a q or a 5.
func (m *Model) handleSynapseKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	p := &m.Synapse
	if p.writing {
		switch msg.String() {
		case "ctrl+c":
			return false, nil
		case "esc":
			p.writing = false
		case "enter":
			if !p.busy {
				return true, m.synapseSend()
			}
		case "backspace":
			if r := []rune(p.draft); len(r) > 0 {
				p.draft = string(r[:len(r)-1])
			}
		default:
			switch msg.Type {
			case tea.KeyRunes:
				p.draft += oneLine(string(msg.Runes))
			case tea.KeySpace:
				p.draft += " "
			}
		}
		return true, nil
	}
	if m.ActiveSheet != types.SheetSynapse || !m.signedIn() {
		return false, nil
	}
	peers := m.synapsePeers()
	switch msg.String() {
	case "j", "down":
		if p.selected < len(peers)-1 {
			p.selected++
		}
	case "k", "up":
		if p.selected > 0 {
			p.selected--
		}
	case "enter", "l", "right":
		if p.selected < len(peers) {
			return true, m.synapseOpen(peers[p.selected])
		}
	case "i", "w":
		if p.peer == "" && p.selected < len(peers) {
			cmd := m.synapseOpen(peers[p.selected])
			p.writing = true
			return true, cmd
		}
		p.writing = p.peer != ""
	case "esc", "h", "left":
		p.peer, p.msgs, p.older = "", nil, false
	case "u", "pgup":
		if p.peer != "" && p.older && len(p.msgs) > 0 {
			return true, m.synapseLoad(p.peer, "", strconv.FormatUint(p.msgs[0].id, 10))
		}
	case "r":
		cmds := []tea.Cmd{m.synapseRefresh()}
		if p.peer != "" {
			cmds = append(cmds, m.synapseLoad(p.peer, "", ""))
		}
		return true, tea.Batch(cmds...)
	default:
		return false, nil
	}
	return true, nil
}

// ─── the wire ────────────────────────────────────────────────────────

// synapsePeers is everyone this person can write to or has written with,
// the latest conversation first.
func (m *Model) synapsePeers() []string {
	me := m.Session.User
	seen := map[string]bool{me: true}
	var out []string
	for _, n := range m.Synapse.people {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	for n := range m.Synapse.chats {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := m.Synapse.chats[out[i]], m.Synapse.chats[out[j]]
		if a.last != b.last {
			return a.last > b.last
		}
		return out[i] < out[j]
	})
	return out
}

func parsePeople(v string) ([]string, error) {
	if v == "" {
		return nil, nil
	}
	names, err := monolink.SplitRecord(v)
	if err != nil {
		return nil, err
	}
	return onlyNames(names), nil // a name is checked, never cleaned
}

// parseMsgs reads a page of MSGS: for each message a header record,
// id|from|to|at|read, then its text as a field of its own.
func parseMsgs(args []string) ([]synapseMsg, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("SYNAPSE sent %d fields for its messages", len(args))
	}
	out := make([]synapseMsg, 0, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		f, err := monolink.SplitRecord(args[i])
		if err != nil || len(f) != 5 {
			return nil, fmt.Errorf("unreadable message %q", args[i])
		}
		id, err := strconv.ParseUint(f[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unreadable message id %q", f[0])
		}
		at, err := time.Parse(time.RFC3339, f[3])
		if err != nil {
			return nil, fmt.Errorf("unreadable message time %q", f[3])
		}
		// Who wrote it is checked; what they wrote is cleaned. SYNAPSE
		// refuses a message with control characters in it, and this is what
		// draws one even if it did not.
		if !marshal.ValidName(f[1]) || !marshal.ValidName(f[2]) {
			return nil, fmt.Errorf("message %s is between names nobody can have", f[0])
		}
		out = append(out, synapseMsg{id: id, from: f[1], to: f[2], at: at, read: f[4] == "ON", text: stripControl(args[i+1])})
	}
	return out, nil
}

// parseChats reads CHATS: peer|unread|last|at for each conversation.
func parseChats(args []string) (map[string]synapseChat, error) {
	out := map[string]synapseChat{}
	for _, a := range args {
		f, err := monolink.SplitRecord(a)
		if err != nil || len(f) != 4 {
			return nil, fmt.Errorf("unreadable conversation %q", a)
		}
		if !marshal.ValidName(f[0]) {
			return nil, fmt.Errorf("a conversation with a name nobody can have")
		}
		unread, _ := strconv.Atoi(f[1])
		last, _ := strconv.ParseUint(f[2], 10, 64)
		at, _ := time.Parse(time.RFC3339, f[3])
		out[f[0]] = synapseChat{unread: unread, last: last, at: at}
	}
	return out, nil
}

// mergePage folds a page into the conversation shown, and says whether
// earlier messages may exist.
func mergePage(have, page []synapseMsg, after, before string, older bool) ([]synapseMsg, bool) {
	switch {
	case before != "":
		older = len(page) == synapsePage
	case after != "":
	default: // the newest page
		if len(have) == 0 || (len(page) == synapsePage && page[0].id > have[len(have)-1].id) {
			return page, len(page) == synapsePage // nothing to join it to, or a gap between
		}
	}
	by := map[uint64]synapseMsg{}
	for _, x := range have {
		by[x.id] = x
	}
	for _, x := range page {
		by[x.id] = x
	}
	out := make([]synapseMsg, 0, len(by))
	for _, x := range by {
		out = append(out, x)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out, older
}

func lastID(msgs []synapseMsg) string {
	if len(msgs) == 0 {
		return ""
	}
	return strconv.FormatUint(msgs[len(msgs)-1].id, 10)
}

func lastUnreadFrom(msgs []synapseMsg, peer string) uint64 {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].from == peer && !msgs[i].read {
			return msgs[i].id
		}
	}
	return 0
}

// splitMessage cuts text into messages SYNAPSE can take: each one field, at
// most monolink.MaxField bytes once escaped. It cuts at a space where one is
// near, else between characters.
func splitMessage(text string) []string {
	var out []string
	text = strings.TrimSpace(text)
	for text != "" {
		cut, size := len(text), 0
		for i, r := range text {
			w := utf8.RuneLen(r)
			if r == '%' || r == ':' {
				w = 3 // escaped as %25 and %3A
			}
			if size+w > monolink.MaxField {
				cut = i
				break
			}
			size += w
		}
		if cut < len(text) {
			if sp := strings.LastIndexByte(text[:cut], ' '); sp > cut/2 {
				cut = sp
			}
		}
		out = append(out, strings.TrimSpace(text[:cut]))
		text = strings.TrimSpace(text[cut:])
	}
	return out
}

// oneLine is typed text as a message can hold it: a frame is one line, so a
// pasted line break becomes a space.
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
}

func synapseDescribe(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "SYNAPSE did not answer"
	}
	return describe(err)
}

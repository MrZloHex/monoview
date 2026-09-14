package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrZloHex/monolink"
	"github.com/MrZloHex/monolink/marshal"
	"monoview/internal/types"
)

// Signing in, and the PEOPLE sheet. MARSHAL keeps people, sessions and
// grants (SPEC §24). This panel signs a person in with its own key, shows
// the hub marshal's ticket for them, sends as MONOVIEW.<person>, and checks
// their grants before sending anything.
//
// The key is Ed25519, in a file sealed with the person's passphrase
// (keyfile.go): opened to sign in, and forgotten once it has signed. The
// session lives as long as this process and is kept nowhere; a restart asks
// for the passphrase again. Every ticket renewed keeps it open.
//
// With nobody signed in it sends nothing but signing in: the hub lets no
// panel act without a person's ticket (SECURITY.txt §5).

const (
	marshalTimeout = 8 * time.Second
	keyTimeout     = 30 * time.Second // argon2id takes a moment first
	ticketEvery    = 4 * time.Minute  // a ticket lasts ten
	minPassLen     = 12
)

type formKind int

const (
	formNone formKind = iota
	formSignIn
	formEnrol
	formRedeem
	formInvite
	formRemoveKey
	formGrant
	formRevoke
)

type formField struct {
	label  string
	value  string
	secret bool
}

// peopleForm is the one form the PEOPLE sheet can have open.
type peopleForm struct {
	kind   formKind
	fields []formField
	focus  int
	user   string // whom a grant or key form is about
	err    string
	busy   bool // waiting for MARSHAL
}

// invitation is a code NEW:INVITE gave, shown until the next one.
type invitation struct {
	Name    string
	Code    string
	Expires time.Time
}

// Answers from MARSHAL, delivered as tea messages: every call to it runs in
// a command, never in Update, and opening a key spends a moment on argon2id.
type (
	sessionMsg struct {
		op  string // "sign in", "enrol", "invitation"
		s   marshal.Session
		err error
	}
	grantsMsg struct {
		user   string
		grants []string
		err    error
	}
	peopleMsg struct {
		users    []string
		grants   map[string][]string
		keys     map[string][]marshal.KeyInfo
		sessions []marshal.SessionInfo
		usersErr error
	}
	inviteMsg struct {
		inv invitation
		err error
	}
	enrollingMsg struct{ on, known bool }
	ticketMsg    struct{ err error }
	peopleOpMsg  struct {
		what string
		err  error
	}
	signedOutMsg struct{}
)

// ReconnectedMsg says the hub connection came back: a new connection, which
// carries no ticket until this panel shows one again.
type ReconnectedMsg struct{}

// RequireSignIn opens the sign-in form when nobody is signed in: the panel
// does nothing else until someone is.
func (m *Model) RequireSignIn() {
	if m.signedIn() {
		return
	}
	m.ActiveSheet = types.SheetPeople
	if m.hasKey() {
		m.openForm(formSignIn, "")
		return
	}
	m.PeopleStatus = "No key at " + m.KeyPath + " yet: [e] enrol as the first person, or [i] take up an invitation."
}

func (m *Model) hasKey() bool {
	if m.KeyPath == "" {
		return false
	}
	_, err := os.Stat(m.KeyPath)
	return err == nil
}

// ticketed shows the hub marshal's ticket for s, which is what lets this
// panel act for its person at all. It happens before the session is adopted
// here, so nothing goes out in their name before the hub knows them.
func ticketed(ctx context.Context, hub *monolink.Client, op string, s marshal.Session, err error) sessionMsg {
	if err == nil {
		_, err = marshal.Ticketed(ctx, hub, s.Token)
	}
	return sessionMsg{op: op, s: s, err: err}
}

// renewTicket keeps the ticket at the hub fresh, and with it the session;
// after a reconnect, puts it back.
func (m *Model) renewTicket() tea.Cmd {
	if !m.signedIn() || time.Since(m.LastTicket) < ticketEvery {
		return nil
	}
	m.LastTicket = time.Now()
	token := m.Session.Token
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		_, err := marshal.Ticketed(ctx, hub, token)
		return ticketMsg{err: err}
	})
}

// ─── the session this panel holds ────────────────────────────────────

func (m *Model) signedIn() bool { return m.Session.Token != "" }

func (m *Model) adopt(s marshal.Session, grants []string) {
	m.Session = s
	m.Grants = grants
	m.deniedLogged = map[string]bool{}
	if m.Hub != nil {
		m.Hub.SetActor(s.User)
	}
}

func (m *Model) dropSession() {
	m.Session = marshal.Session{}
	m.SignedInAt = time.Time{}
	m.Grants = nil
	m.People, m.PeopleGrants, m.PeopleKeys, m.PeopleSessions = nil, nil, nil, nil
	m.Invitation = invitation{}
	m.Synapse = synapseState{} // their messages are not the next person's
	m.LastTicket = time.Time{}
	if m.Hub != nil {
		m.Hub.SetActor("")
	}
}

// permitted reports whether the person signed in here may send this, and
// logs the first refusal of each action. MARSHAL judges its own requests,
// and signing in goes there. Otherwise nothing is sent with nobody signed
// in; PING is allowed to whoever is — it asks nothing of a node but that it
// exists.
func (m *Model) permitted(to, verb, noun string) bool {
	if strings.EqualFold(to, marshal.Node) {
		return true
	}
	action := marshal.Action(strings.ToUpper(to), verb, noun)
	switch {
	case !m.signedIn():
		m.refusedOnce("sign in", "nothing is sent until someone signs in")
		return false
	case verb == monolink.VerbPing, marshal.Allowed(m.Grants, action):
		return true
	}
	m.refusedOnce(action, "not permitted: "+action)
	return false
}

func (m *Model) refusedOnce(key, text string) {
	if m.deniedLogged == nil {
		m.deniedLogged = map[string]bool{}
	}
	if !m.deniedLogged[key] {
		m.deniedLogged[key] = true
		m.peopleLog("WARN", text)
	}
}

func (m *Model) peopleLog(level, text string) {
	m.Logs = append([]types.LogEntry{{Time: time.Now(), Level: level, Source: "MONOVIEW", Message: text}}, m.Logs...)
	if len(m.Logs) > 50 {
		m.Logs = m.Logs[:50]
	}
}

// handleMarshalPub notes MARSHAL's ENROLLING property, published whenever
// it changes, so the sheet can say the first person is awaited.
func (m *Model) handleMarshalPub(msg monolink.Message) {
	if msg.Verb == monolink.VerbPub && strings.EqualFold(msg.From, marshal.Node) &&
		msg.Noun == "ENROLLING" && len(msg.Args) > 0 {
		m.MarshalEnrolling = msg.Args[0] == "ON"
	}
}

// ─── asking MARSHAL ──────────────────────────────────────────────────

func (m *Model) ask(f func(ctx context.Context, hub *monolink.Client) tea.Msg) tea.Cmd {
	return m.askWithin(marshalTimeout, f)
}

func (m *Model) askWithin(d time.Duration, f func(ctx context.Context, hub *monolink.Client) tea.Msg) tea.Cmd {
	hub := m.Hub
	if hub == nil {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()
		return f(ctx, hub)
	}
}

// op runs an administrative call and reports it as a peopleOpMsg.
func (m *Model) op(what string, f func(ctx context.Context, hub *monolink.Client) error) tea.Cmd {
	m.PeopleForm.busy = true
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		return peopleOpMsg{what: what, err: f(ctx, hub)}
	})
}

func (m *Model) grantsCmd(user string) tea.Cmd {
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		g, err := marshal.Grants(ctx, hub, user)
		return grantsMsg{user: user, grants: cleanAll(g), err: err}
	})
}

// refreshPeople asks whether MARSHAL awaits its first person and, for
// whoever is signed in, everything their grants let them see — their own
// keys always.
func (m *Model) refreshPeople() tea.Cmd {
	enrolling := m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		r, err := hub.RequestDialect(ctx, monolink.V2, marshal.Node, monolink.VerbGet, "ENROLLING")
		return enrollingMsg{on: err == nil && len(r.Args) > 0 && r.Args[0] == "ON", known: err == nil}
	})
	if !m.signedIn() {
		return enrolling
	}
	self := m.Session.User
	canUsers := marshal.Allowed(m.Grants, "MARSHAL.GET.USERS")
	canKeys := marshal.Allowed(m.Grants, "MARSHAL.GET.KEYS")
	canSessions := marshal.Allowed(m.Grants, "MARSHAL.GET.SESSIONS")
	list := m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		out := peopleMsg{users: []string{self}, grants: map[string][]string{}, keys: map[string][]marshal.KeyInfo{}}
		if canUsers {
			if u, err := marshal.Users(ctx, hub); err != nil {
				out.usersErr = err
			} else {
				out.users = onlyNames(u)
			}
		}
		for _, u := range out.users {
			if g, err := marshal.Grants(ctx, hub, u); err == nil {
				out.grants[u] = cleanAll(g)
			}
			if u == self || canKeys {
				if k, err := marshal.Keys(ctx, hub, u); err == nil {
					out.keys[u] = cleanKeys(k)
				}
			}
		}
		if canSessions {
			s, _ := marshal.Sessions(ctx, hub)
			out.sessions = cleanSessions(s)
		}
		return out
	})
	return tea.Batch(enrolling, list)
}

// onlyNames keeps the names a person can have. A name is checked, never
// cleaned: stripping an escape out of one would make it another person's.
func onlyNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, n := range names {
		if marshal.ValidName(n) {
			out = append(out, n)
		}
	}
	return out
}

// cleanKeys takes the control characters out of what a person called a key.
// marshal keeps only printable labels; this is what draws them even if it
// did not.
func cleanKeys(keys []marshal.KeyInfo) []marshal.KeyInfo {
	out := make([]marshal.KeyInfo, len(keys))
	for i, k := range keys {
		k.Ref, k.Kind, k.Label = stripControl(k.Ref), stripControl(k.Kind), stripControl(k.Label)
		out[i] = k
	}
	return out
}

// cleanSessions does the same for the sessions list.
func cleanSessions(ss []marshal.SessionInfo) []marshal.SessionInfo {
	out := make([]marshal.SessionInfo, len(ss))
	for i, s := range ss {
		s.User, s.Panel = stripControl(s.User), stripControl(s.Panel)
		out[i] = s
	}
	return out
}

// describe turns MARSHAL's answer into a line for a person. The detail in it
// is another node's text, come back outside the inbox and so past plain: the
// control characters go here, before it is ever drawn.
func describe(err error) string { return stripControl(detail(err)) }

func detail(err error) string {
	var re *monolink.ReplyError
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "MARSHAL did not answer"
	case errors.As(err, &re):
		switch re.Code {
		case monolink.CodeDenied:
			if re.Detail != "" {
				return "not permitted: " + re.Detail
			}
			return "not accepted"
		case monolink.CodeBusy:
			return "too many wrong codes; try again in " + re.Detail + " s"
		case monolink.CodeNAC, monolink.CodeState:
			if re.Detail != "" {
				return re.Detail
			}
		case monolink.CodeArg:
			return "not accepted: " + re.Detail
		}
		return strings.TrimSpace(re.Code + " " + re.Detail)
	}
	return err.Error()
}

func (m *Model) handlePeopleMsg(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case sessionMsg:
		m.PeopleForm.busy = false
		if msg.err != nil {
			m.PeopleForm.err = describe(msg.err)
			return nil
		}
		m.adopt(msg.s, nil)
		m.SignedInAt = time.Now()
		m.LastTicket = time.Now()
		m.loadSheets()
		m.closeForm()
		m.PeopleStatus = "Signed in as " + msg.s.User + "."
		m.peopleLog("INFO", "signed in as "+msg.s.User+" ("+msg.op+")")
		return tea.Batch(m.grantsCmd(msg.s.User), m.refreshPeople())

	case grantsMsg:
		if msg.err == nil && msg.user == m.Session.User {
			m.Grants = msg.grants
			m.deniedLogged = map[string]bool{}
			return m.synapseRefresh() // now it is known whether SYNAPSE may be asked
		}

	case peopleMsg:
		m.People, m.PeopleGrants, m.PeopleKeys, m.PeopleSessions = msg.users, msg.grants, msg.keys, msg.sessions
		m.PeopleNote = ""
		if msg.usersErr != nil {
			m.PeopleNote = describe(msg.usersErr)
		}
		if m.PeopleSelected >= len(m.People) {
			m.PeopleSelected = max(len(m.People)-1, 0)
		}

	case inviteMsg:
		m.PeopleForm.busy = false
		if msg.err != nil {
			m.PeopleForm.err = describe(msg.err)
			return nil
		}
		m.closeForm()
		m.Invitation = msg.inv
		m.peopleLog("INFO", "invited "+msg.inv.Name)

	case enrollingMsg:
		if msg.known {
			m.MarshalEnrolling = msg.on
		}

	case ticketMsg:
		var re *monolink.ReplyError
		switch {
		case errors.As(msg.err, &re):
			m.peopleLog("WARN", "the ticket was refused: "+describe(msg.err))
			m.dropSession()
			m.RequireSignIn()
			m.PeopleStatus = "Your session has ended. Sign in again."
		case msg.err != nil:
			m.peopleLog("WARN", "ticket not renewed: "+msg.err.Error())
			m.LastTicket = time.Now().Add(-ticketEvery + 30*time.Second) // try again shortly
		}

	case peopleOpMsg:
		m.PeopleForm.busy = false
		if msg.err != nil {
			if m.PeopleForm.kind != formNone {
				m.PeopleForm.err = describe(msg.err)
			} else {
				m.PeopleStatus = msg.what + " failed: " + describe(msg.err)
			}
			return nil
		}
		m.closeForm()
		m.PeopleStatus = msg.what + "."
		return tea.Batch(m.grantsCmd(m.Session.User), m.refreshPeople())
	}
	return nil
}

// ─── the PEOPLE sheet ────────────────────────────────────────────────

func (m *Model) selectedPerson() string {
	if m.PeopleSelected >= 0 && m.PeopleSelected < len(m.People) {
		return m.People[m.PeopleSelected]
	}
	return ""
}

func (m *Model) openForm(kind formKind, user string) {
	f := peopleForm{kind: kind, user: user}
	switch kind {
	case formSignIn:
		f.fields = []formField{{label: "Passphrase", secret: true}}
	case formEnrol, formRedeem:
		f.fields = []formField{{label: "Code"}, {label: "Name"},
			{label: "Passphrase", secret: true}, {label: "Again", secret: true}}
	case formInvite:
		f.fields = []formField{{label: "Name"}}
	case formRemoveKey:
		f.fields = []formField{{label: "Key"}}
	case formGrant, formRevoke:
		f.fields = []formField{{label: "Grant"}}
	}
	m.PeopleForm = f
	m.PeopleConfirm = ""
	m.PeopleStatus = ""
}

// closeForm also drops whatever passphrase was typed into it.
func (m *Model) closeForm() { m.PeopleForm = peopleForm{} }

// handlePeopleKeys serves the PEOPLE sheet and, wherever it is open, its
// form — which takes every key, so a name may contain a q or a 5.
func (m *Model) handlePeopleKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	if m.PeopleForm.kind != formNone {
		return m.handlePeopleFormKeys(msg)
	}
	if m.ActiveSheet != types.SheetPeople {
		return false, nil
	}
	if who := m.PeopleConfirm; who != "" {
		m.PeopleConfirm = ""
		if msg.String() != "y" {
			m.PeopleStatus = "Kept " + who + "."
			return true, nil
		}
		token := m.Session.Token
		return true, m.op("Removed "+who, func(ctx context.Context, hub *monolink.Client) error {
			return marshal.RemoveUser(ctx, hub, token, who)
		})
	}

	switch msg.String() {
	case "s":
		if !m.hasKey() {
			m.RequireSignIn()
			return true, nil
		}
		m.openForm(formSignIn, "")
		return true, nil
	case "e":
		m.openForm(formEnrol, "")
		return true, nil
	case "i":
		m.openForm(formRedeem, "")
		return true, nil
	case "r":
		return true, m.refreshPeople()
	case "j", "down":
		if m.PeopleSelected < len(m.People)-1 {
			m.PeopleSelected++
		}
		return true, nil
	case "k", "up":
		if m.PeopleSelected > 0 {
			m.PeopleSelected--
		}
		return true, nil
	}
	if !m.signedIn() {
		return false, nil
	}
	switch msg.String() {
	case "n", "K", "g", "x", "D":
		// marshal takes a change to people only from a session signed in
		// within five minutes: better to sign in again now than be refused
		// after filling in the form.
		if time.Since(m.SignedInAt) > marshal.FreshSignIn-15*time.Second {
			m.openForm(formSignIn, "")
			m.PeopleStatus = "A change to people, keys or grants needs a sign-in within five minutes: sign in again, then repeat it."
			return true, nil
		}
	}
	switch msg.String() {
	case "o":
		token, who := m.Session.Token, m.Session.User
		m.dropSession()
		m.PeopleStatus = "Signed out."
		m.peopleLog("INFO", "signed out "+who)
		return true, m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
			marshal.SignOut(ctx, hub, token)
			marshal.DropTicket(ctx, hub)
			return signedOutMsg{}
		})
	case "n":
		m.openForm(formInvite, "")
	case "K":
		if u := m.selectedPerson(); u != "" {
			m.openForm(formRemoveKey, u)
		}
	case "g", "x":
		if u := m.selectedPerson(); u != "" {
			kind := formGrant
			if msg.String() == "x" {
				kind = formRevoke
			}
			m.openForm(kind, u)
		}
	case "D":
		m.PeopleConfirm = m.selectedPerson()
	default:
		return false, nil
	}
	return true, nil
}

func (m *Model) handlePeopleFormKeys(msg tea.KeyMsg) (bool, tea.Cmd) {
	f := &m.PeopleForm
	switch msg.String() {
	case "ctrl+c":
		return false, nil
	case "esc":
		m.closeForm()
		return true, nil
	case "tab", "down":
		f.focus = (f.focus + 1) % len(f.fields)
	case "shift+tab", "up":
		f.focus = (f.focus + len(f.fields) - 1) % len(f.fields)
	case "enter":
		if f.busy {
			return true, nil
		}
		if f.focus < len(f.fields)-1 {
			f.focus++
			return true, nil
		}
		return true, m.submitForm()
	case "backspace":
		if r := []rune(f.fields[f.focus].value); len(r) > 0 {
			f.fields[f.focus].value = string(r[:len(r)-1])
		}
	default:
		switch msg.Type {
		case tea.KeyRunes:
			f.fields[f.focus].value += string(msg.Runes)
		case tea.KeySpace:
			f.fields[f.focus].value += " "
		}
	}
	f.err = ""
	return true, nil
}

func (m *Model) submitForm() tea.Cmd {
	f := &m.PeopleForm
	if m.Hub == nil {
		f.err = "not connected to the hub"
		return nil
	}
	val := func(i int) string {
		if f.fields[i].secret {
			return f.fields[i].value
		}
		return strings.TrimSpace(f.fields[i].value)
	}
	name := func(i int) (string, bool) {
		n := strings.ToLower(val(i))
		if !marshal.ValidName(n) {
			f.err = "a name is a-z, 0-9, _ and -, starting with a letter or digit"
			return "", false
		}
		return n, true
	}
	newPass := func(a, b int) (string, bool) {
		s := val(a)
		if len([]rune(s)) < minPassLen {
			f.err = fmt.Sprintf("a passphrase needs at least %d characters", minPassLen)
			return "", false
		}
		if s != val(b) {
			f.err = "the two passphrases differ"
			return "", false
		}
		return s, true
	}
	path := m.KeyPath
	token := m.Session.Token // the session a change is made in

	switch f.kind {
	case formSignIn:
		pass := val(0)
		f.busy = true
		return m.askWithin(keyTimeout, func(ctx context.Context, hub *monolink.Client) tea.Msg {
			s, err := signInWithKey(ctx, hub, path, pass)
			return ticketed(ctx, hub, "sign in", s, err)
		})

	case formEnrol, formRedeem:
		code := val(0)
		n, ok := name(1)
		if !ok {
			return nil
		}
		pass, ok := newPass(2, 3)
		if !ok {
			return nil
		}
		invited := f.kind == formRedeem
		f.busy = true
		return m.askWithin(keyTimeout, func(ctx context.Context, hub *monolink.Client) tea.Msg {
			op := "enrol"
			if invited {
				op = "invitation"
			}
			s, err := makeKey(ctx, hub, path, code, n, pass, invited)
			return ticketed(ctx, hub, op, s, err)
		})

	case formInvite:
		n, ok := name(0)
		if !ok {
			return nil
		}
		f.busy = true
		return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
			code, exp, err := marshal.Invite(ctx, hub, token, n)
			return inviteMsg{inv: invitation{Name: n, Code: stripControl(code), Expires: exp}, err: err}
		})

	case formRemoveKey:
		user, which := f.user, val(0)
		ref := ""
		for _, k := range m.PeopleKeys[user] {
			if which != "" && (k.Label == which || strings.HasPrefix(k.Ref, which)) {
				ref = k.Ref
				break
			}
		}
		if ref == "" {
			f.err = "no such key of " + user + ": its label, or its ref"
			return nil
		}
		return m.op("Removed a key of "+user, func(ctx context.Context, hub *monolink.Client) error {
			return marshal.RemoveKey(ctx, hub, token, user, ref)
		})

	case formGrant, formRevoke:
		p := val(0)
		if !marshal.ValidPattern(p) {
			f.err = "a grant is like VERTEX.* or UKAZ.DO.PRINT.AGENDA"
			return nil
		}
		user := f.user
		if f.kind == formGrant {
			return m.op("Granted "+p+" to "+user, func(ctx context.Context, hub *monolink.Client) error {
				return marshal.Grant(ctx, hub, token, user, p)
			})
		}
		return m.op("Revoked "+p+" from "+user, func(ctx context.Context, hub *monolink.Client) error {
			return marshal.Revoke(ctx, hub, token, user, p)
		})
	}
	return nil
}

// signInWithKey opens the key at path and signs its person in with it.
func signInWithKey(ctx context.Context, hub *monolink.Client, path, pass string) (marshal.Session, error) {
	kf, err := LoadKeyFile(path)
	if err != nil {
		return marshal.Session{}, err
	}
	key, err := kf.Open(pass)
	if err != nil {
		return marshal.Session{}, err
	}
	defer clear(key)
	return marshal.SignInKey(ctx, hub, kf.Person, kf.ID, key)
}

// makeKey makes this panel's key for name, seals it in a file at path, and
// has marshal take it — by the enrolment code, or an invitation — once the
// key has signed marshal's challenge, which shows it is this panel's. The
// file is written first, so that a key marshal took is never one nobody
// holds; only if marshal refuses it does the file go again.
func makeKey(ctx context.Context, hub *monolink.Client, path, code, name, pass string, invited bool) (marshal.Session, error) {
	if path == "" {
		return marshal.Session{}, errors.New("no --key path to keep a key at")
	}
	kf, key, cred, err := NewKeyFile(name, keyLabel(), pass)
	if err != nil {
		return marshal.Session{}, err
	}
	defer clear(key)
	if err := kf.Save(path); err != nil {
		return marshal.Session{}, err
	}
	prove := marshal.KeyProof(key, name, hub.NodeID())
	var s marshal.Session
	if invited {
		s, err = marshal.Redeem(ctx, hub, code, name, cred, prove)
	} else {
		s, err = marshal.Enrol(ctx, hub, code, name, cred, prove)
	}
	var re *monolink.ReplyError
	switch {
	case errors.As(err, &re):
		os.Remove(path)
	case err != nil:
		err = fmt.Errorf("%s — the key is kept at %s; try [s] signing in with it", describe(err), path)
	}
	return s, err
}

// keyLabel is what marshal lists this panel's key as: monoview@<host>.
func keyLabel() string {
	host, _ := os.Hostname()
	host = strings.Map(func(r rune) rune {
		if r < 0x80 && r != ':' && r != '|' && r != '%' && r > ' ' {
			return r
		}
		return -1
	}, host)
	l := "monoview@" + host
	return l[:min(len(l), marshal.MaxLabel)]
}

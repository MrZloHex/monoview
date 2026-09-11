package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrZloHex/monolink"
	"github.com/MrZloHex/monolink/marshal"
	"monoview/internal/types"
)

// Signing in, and the PEOPLE sheet. MARSHAL keeps people, sessions and
// grants (SPEC §24). This panel signs a person in, sends as
// MONOVIEW.<person>, and checks their grants before sending anything.
//
// Nobody signed in is the panel as it always was: its owner's, unqualified.

const (
	marshalTimeout = 8 * time.Second
	keepAliveEvery = time.Hour
	minSecretLen   = 8
)

type formKind int

const (
	formNone formKind = iota
	formSignIn
	formEnrol
	formOwnSecret
	formNewUser
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
	user   string // whom a grant form is about
	err    string
	busy   bool // waiting for MARSHAL
}

// Answers from MARSHAL, delivered as tea messages: every call to it runs in
// a command, never in Update, and a sign-in spends a moment on argon2id.
type (
	sessionMsg struct {
		op  string // "sign in", "enrol", "resume"
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
		sessions []marshal.SessionInfo
		usersErr error
	}
	enrollingMsg struct{ on, known bool }
	peopleOpMsg  struct {
		what string
		err  error
	}
	signedOutMsg struct{}
)

// ─── the session this panel holds ────────────────────────────────────

func (m *Model) signedIn() bool { return m.Session.Token != "" }

type savedSession struct {
	Token   string    `json:"token"`
	User    string    `json:"user"`
	Expires time.Time `json:"expires"`
	Grants  []string  `json:"grants"`
}

// RestoreSession picks up the session this panel held when it last ran, so
// a restart signs nobody out. Init then asks MARSHAL whether it stands.
func (m *Model) RestoreSession() {
	if m.SessionPath == "" {
		return
	}
	b, err := os.ReadFile(m.SessionPath)
	if err != nil {
		return
	}
	var s savedSession
	if json.Unmarshal(b, &s) != nil || s.Token == "" || !time.Now().Before(s.Expires) {
		m.dropSession()
		return
	}
	m.adopt(marshal.Session{Token: s.Token, User: s.User, Expires: s.Expires}, s.Grants)
	m.LastKeepAlive = time.Now()
}

func (m *Model) adopt(s marshal.Session, grants []string) {
	m.Session = s
	m.Grants = grants
	m.deniedLogged = map[string]bool{}
	if m.Hub != nil {
		m.Hub.SetActor(s.User)
	}
}

// saveSession keeps the session, and the grants last seen, for the next
// run — readable by its owner only, since the token is the session.
func (m *Model) saveSession() {
	if m.SessionPath == "" || !m.signedIn() {
		return
	}
	b, _ := json.MarshalIndent(savedSession{m.Session.Token, m.Session.User, m.Session.Expires, m.Grants}, "", "  ")
	tmp, err := os.CreateTemp(filepath.Dir(m.SessionPath), ".monoview-session-*")
	if err != nil {
		m.peopleLog("WARN", "cannot keep the session: "+err.Error())
		return
	}
	name := tmp.Name()
	defer os.Remove(name) // no-op once renamed
	_, werr := tmp.Write(append(b, '\n'))
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr); err != nil {
		m.peopleLog("WARN", "cannot keep the session: "+err.Error())
		return
	}
	if err := os.Rename(name, m.SessionPath); err != nil {
		m.peopleLog("WARN", "cannot keep the session: "+err.Error())
	}
}

func (m *Model) dropSession() {
	m.Session = marshal.Session{}
	m.Grants = nil
	m.People, m.PeopleGrants, m.PeopleSessions = nil, nil, nil
	if m.Hub != nil {
		m.Hub.SetActor("")
	}
	if m.SessionPath != "" {
		os.Remove(m.SessionPath)
	}
}

// permitted reports whether the person signed in here may send this, and
// logs the first refusal of each action. With nobody signed in the panel is
// its owner's, as before marshal. PING is always allowed: it asks nothing
// of a node but that it exists. MARSHAL judges its own requests.
func (m *Model) permitted(to, verb, noun string) bool {
	if !m.signedIn() || verb == monolink.VerbPing || strings.EqualFold(to, marshal.Node) {
		return true
	}
	action := marshal.Action(strings.ToUpper(to), verb, noun)
	if marshal.Allowed(m.Grants, action) {
		return true
	}
	if m.deniedLogged == nil {
		m.deniedLogged = map[string]bool{}
	}
	if !m.deniedLogged[action] {
		m.deniedLogged[action] = true
		m.peopleLog("WARN", "not permitted: "+action)
	}
	return false
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
	hub := m.Hub
	if hub == nil {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), marshalTimeout)
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

func (m *Model) resumeCmd() tea.Cmd {
	token := m.Session.Token
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		s, err := marshal.Resume(ctx, hub, token)
		return sessionMsg{op: "resume", s: s, err: err}
	})
}

// keepAlive extends the session once an hour while this panel runs.
func (m *Model) keepAlive() tea.Cmd {
	if !m.signedIn() || time.Since(m.LastKeepAlive) < keepAliveEvery {
		return nil
	}
	m.LastKeepAlive = time.Now()
	return m.resumeCmd()
}

func (m *Model) grantsCmd(user string) tea.Cmd {
	return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		g, err := marshal.Grants(ctx, hub, user)
		return grantsMsg{user: user, grants: g, err: err}
	})
}

// refreshPeople asks whether MARSHAL awaits its first person and, for
// whoever is signed in, everything their grants let them see.
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
	canSessions := marshal.Allowed(m.Grants, "MARSHAL.GET.SESSIONS")
	list := m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
		out := peopleMsg{users: []string{self}, grants: map[string][]string{}}
		if canUsers {
			if u, err := marshal.Users(ctx, hub); err != nil {
				out.usersErr = err
			} else {
				out.users = u
			}
		}
		for _, u := range out.users {
			if g, err := marshal.Grants(ctx, hub, u); err == nil {
				out.grants[u] = g
			}
		}
		if canSessions {
			out.sessions, _ = marshal.Sessions(ctx, hub)
		}
		return out
	})
	return tea.Batch(enrolling, list)
}

// describe turns MARSHAL's answer into a line for a person.
func describe(err error) string {
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
			return "wrong name, secret or code"
		case monolink.CodeBusy:
			return "locked out; try again in " + re.Detail + " s"
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
			if msg.op != "resume" {
				m.PeopleForm.err = describe(msg.err)
				return nil
			}
			var re *monolink.ReplyError
			if errors.As(msg.err, &re) && re.Code == monolink.CodeNAC {
				m.peopleLog("WARN", "session of "+m.Session.User+" has ended")
				m.dropSession()
				m.PeopleStatus = "Your session has ended. Sign in again."
			} else {
				// MARSHAL down is not the bubble locked (SPEC §39): the session
				// stands until it expires.
				m.PeopleStatus = "MARSHAL did not answer; keeping the session until " +
					m.Session.Expires.Local().Format("2006-01-02 15:04")
			}
			return nil
		}
		grants := m.Grants
		if msg.op != "resume" || msg.s.User != m.Session.User {
			grants = nil
		}
		m.adopt(msg.s, grants)
		m.LastKeepAlive = time.Now()
		if msg.op != "resume" {
			m.closeForm()
			m.PeopleStatus = "Signed in as " + msg.s.User + "."
			m.peopleLog("INFO", "signed in as "+msg.s.User)
		}
		m.saveSession()
		return tea.Batch(m.grantsCmd(msg.s.User), m.refreshPeople())

	case grantsMsg:
		if msg.err == nil && msg.user == m.Session.User {
			m.Grants = msg.grants
			m.deniedLogged = map[string]bool{}
			m.saveSession()
		}

	case peopleMsg:
		m.People, m.PeopleGrants, m.PeopleSessions = msg.users, msg.grants, msg.sessions
		m.PeopleNote = ""
		if msg.usersErr != nil {
			m.PeopleNote = describe(msg.usersErr)
		}
		if m.PeopleSelected >= len(m.People) {
			m.PeopleSelected = max(len(m.People)-1, 0)
		}

	case enrollingMsg:
		if msg.known {
			m.MarshalEnrolling = msg.on
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
		f.fields = []formField{{label: "Name"}, {label: "Secret", secret: true}}
	case formEnrol:
		f.fields = []formField{{label: "Code"}, {label: "Name"},
			{label: "Secret", secret: true}, {label: "Again", secret: true}}
	case formOwnSecret:
		f.fields = []formField{{label: "New secret", secret: true}, {label: "Again", secret: true}}
	case formNewUser:
		f.fields = []formField{{label: "Name"}, {label: "Secret", secret: true}, {label: "Again", secret: true}}
	case formGrant, formRevoke:
		f.fields = []formField{{label: "Grant"}}
	}
	m.PeopleForm = f
	m.PeopleConfirm = ""
	m.PeopleStatus = ""
}

// closeForm also drops whatever secret was typed into it.
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
		return true, m.op("Removed "+who, func(ctx context.Context, hub *monolink.Client) error {
			return marshal.RemoveUser(ctx, hub, who)
		})
	}

	switch msg.String() {
	case "s":
		m.openForm(formSignIn, "")
		return true, nil
	case "e":
		m.openForm(formEnrol, "")
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
	case "o":
		token, who := m.Session.Token, m.Session.User
		m.dropSession()
		m.PeopleStatus = "Signed out."
		m.peopleLog("INFO", "signed out "+who)
		return true, m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
			marshal.SignOut(ctx, hub, token)
			return signedOutMsg{}
		})
	case "p":
		m.openForm(formOwnSecret, m.Session.User)
	case "n":
		m.openForm(formNewUser, "")
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
	newSecret := func(a, b int) (string, bool) {
		s := val(a)
		if len([]rune(s)) < minSecretLen {
			f.err = fmt.Sprintf("a secret needs at least %d characters", minSecretLen)
			return "", false
		}
		if s != val(b) {
			f.err = "the two secrets differ"
			return "", false
		}
		return s, true
	}

	switch f.kind {
	case formSignIn:
		n, ok := name(0)
		if !ok {
			return nil
		}
		secret := val(1)
		f.busy = true
		return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
			s, err := marshal.SignIn(ctx, hub, n, secret)
			return sessionMsg{op: "sign in", s: s, err: err}
		})

	case formEnrol:
		code := val(0)
		n, ok := name(1)
		if !ok {
			return nil
		}
		secret, ok := newSecret(2, 3)
		if !ok {
			return nil
		}
		f.busy = true
		return m.ask(func(ctx context.Context, hub *monolink.Client) tea.Msg {
			s, err := marshal.Enrol(ctx, hub, code, n, secret)
			return sessionMsg{op: "enrol", s: s, err: err}
		})

	case formOwnSecret:
		secret, ok := newSecret(0, 1)
		if !ok {
			return nil
		}
		user := m.Session.User
		return m.op("Secret changed", func(ctx context.Context, hub *monolink.Client) error {
			return marshal.SetSecret(ctx, hub, user, secret)
		})

	case formNewUser:
		n, ok := name(0)
		if !ok {
			return nil
		}
		secret, ok := newSecret(1, 2)
		if !ok {
			return nil
		}
		return m.op("Added "+n, func(ctx context.Context, hub *monolink.Client) error {
			return marshal.NewUser(ctx, hub, n, secret)
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
				return marshal.Grant(ctx, hub, user, p)
			})
		}
		return m.op("Revoked "+p+" from "+user, func(ctx context.Context, hub *monolink.Client) error {
			return marshal.Revoke(ctx, hub, user, p)
		})
	}
	return nil
}

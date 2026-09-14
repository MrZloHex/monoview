package app

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"monoview/internal/ui"
)

const peopleBoxWidth = 50

func (m Model) renderPeople() string {
	left := m.renderThisPanel()
	if m.Invitation.Code != "" {
		left += "\n" + m.renderInvitation()
	}
	if m.PeopleForm.kind != formNone {
		left += "\n" + m.renderPeopleForm()
	}
	if m.PeopleStatus != "" {
		left += "\n\n" + ui.Dim.Render("  "+m.PeopleStatus)
	}
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "    ", m.renderPeopleLists())
	return ui.IndentLines(content, "  ")
}

func (m Model) renderThisPanel() string {
	var lines []string
	lines = append(lines, "")
	if m.signedIn() {
		grants := strings.Join(m.Grants, "  ")
		if grants == "" {
			grants = "nothing yet"
		}
		lines = append(lines,
			"  "+ui.Label.Render("Signed in as ")+ui.Accent.Render(m.Session.User),
			"  "+ui.Label.Render("Until        ")+ui.Value.Render(m.Session.Expires.Local().Format("2006-01-02 15:04")),
			"  "+ui.Label.Render("May          ")+ui.Value.Render(ui.TruncateString(grants, peopleBoxWidth-19)),
		)
	} else {
		lines = append(lines,
			"  "+ui.Dim.Render("Nobody is signed in at this panel;"),
			"  "+ui.Dim.Render("it sends nothing until someone is."),
		)
		if m.MarshalEnrolling {
			lines = append(lines, "",
				"  "+ui.Warning.Render("MARSHAL is waiting for its first person."),
				"  "+ui.Dim.Render("The code is on UKAZ's paper. Press [e]."),
			)
		}
	}
	lines = append(lines, "")
	box := ui.NewBox(peopleBoxWidth).WithTitle(" THIS PANEL ")
	return ui.Title.Render("▌SIGNED IN") + "\n\n" + box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderInvitation() string {
	inv := m.Invitation
	lines := []string{"",
		"  " + ui.Label.Render("For   ") + ui.Accent.Render(inv.Name),
		"  " + ui.Label.Render("Code  ") + ui.Warning.Render(inv.Code),
		"  " + ui.Label.Render("Until ") + ui.Value.Render(inv.Expires.Local().Format("2006-01-02 15:04")),
		"",
		ui.Dim.Render("  On their phone: https://monolith-system.net,"),
		ui.Dim.Render("  “I have an invitation”, the name and the code."),
		ui.Dim.Render("  At a monoview: [i]. It works once."),
		"",
	}
	box := ui.NewBox(peopleBoxWidth).WithBorderColor(ui.GruvAqua).WithTitle(" INVITATION ")
	return box.Render(strings.Join(lines, "\n"))
}

var formTitles = map[formKind]string{
	formSignIn:    " SIGN IN ",
	formEnrol:     " FIRST PERSON ",
	formRedeem:    " INVITATION ",
	formInvite:    " INVITE ",
	formRemoveKey: " REMOVE A KEY ",
	formGrant:     " GRANT ",
	formRevoke:    " REVOKE ",
}

func (m Model) renderPeopleForm() string {
	f := m.PeopleForm
	lines := []string{""}
	switch f.kind {
	case formGrant, formRevoke, formRemoveKey:
		lines = append(lines, "  "+ui.Label.Render("Person: ")+ui.Accent.Render(f.user), "")
	}
	for i, fld := range f.fields {
		val := fld.value
		if fld.secret {
			val = strings.Repeat("•", utf8.RuneCountInString(val))
		}
		line := ui.Label.Render(fmt.Sprintf("  %-11s ", fld.label+":")) + ui.Value.Render(val)
		if i == f.focus {
			line += ui.Dim.Render("▌")
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	switch f.kind {
	case formSignIn:
		lines = append(lines, ui.Dim.Render("  The passphrase of the key at "+ui.TruncateString(m.KeyPath, peopleBoxWidth-34)+"."))
	case formEnrol:
		lines = append(lines, ui.Dim.Render("  The code MARSHAL printed on UKAZ. This panel"),
			ui.Dim.Render("  makes a key, sealed with the passphrase; you"),
			ui.Dim.Render("  become the first person, allowed everything."))
	case formRedeem:
		lines = append(lines, ui.Dim.Render("  The code you were given. This panel makes a"),
			ui.Dim.Render("  key for you, sealed with the passphrase."))
	case formInvite:
		lines = append(lines, ui.Dim.Render("  Someone new — or yourself, for another device."),
			ui.Dim.Render("  The code works once, within the hour."))
	case formRemoveKey:
		lines = append(lines, ui.Dim.Render("  Its label or its ref, as listed. What it"),
			ui.Dim.Render("  signed in ends; the last key stays."))
	case formGrant, formRevoke:
		lines = append(lines, ui.Dim.Render("  VERTEX.*   UKAZ.DO.PRINT.*   GOVERNOR.GET.*   *"))
	}
	switch {
	case f.busy && (f.kind == formSignIn || f.kind == formEnrol || f.kind == formRedeem):
		lines = append(lines, "  "+ui.Dim.Render("opening the key, asking MARSHAL…"))
	case f.busy:
		lines = append(lines, "  "+ui.Dim.Render("asking MARSHAL…"))
	case f.err != "":
		lines = append(lines, "  "+ui.Offline.Render(f.err))
	}
	lines = append(lines, ui.Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
	box := ui.NewBox(peopleBoxWidth).WithBorderColor(ui.GruvAqua).WithTitle(formTitles[f.kind])
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderPeopleLists() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("▌PEOPLE") + "\n\n")
	if !m.signedIn() {
		b.WriteString(ui.Dim.Render("Sign in to see the people of the bubble."))
		return b.String()
	}
	if m.PeopleNote != "" {
		b.WriteString(ui.Dim.Render(m.PeopleNote) + "\n")
	}
	for i, u := range m.People {
		cursor, name := "  ", fmt.Sprintf("%-12s", u)
		if i == m.PeopleSelected {
			cursor, name = ui.Accent.Render("▸ "), ui.Accent.Render(name)
		} else {
			name = ui.Value.Render(name)
		}
		grants := ui.Dim.Render("—")
		if g, known := m.PeopleGrants[u]; known {
			if len(g) == 0 {
				grants = ui.Dim.Render("nothing")
			} else {
				grants = ui.Label.Render(strings.Join(g, "  "))
			}
		}
		you := ""
		if u == m.Session.User {
			you = ui.Dim.Render("  (you)")
		}
		b.WriteString(cursor + name + " " + grants + you + "\n")
		if keys, known := m.PeopleKeys[u]; known {
			names := make([]string, len(keys))
			for j, k := range keys {
				names[j] = k.Label
				if names[j] == "" {
					names[j] = k.Ref
				}
				if k.Kind != "webauthn" {
					names[j] += " (panel)"
				}
			}
			b.WriteString(ui.Dim.Render("               keys: "+strings.Join(names, " · ")) + "\n")
		}
	}
	if m.PeopleConfirm != "" {
		b.WriteString("\n" + ui.Warning.Render("Remove "+m.PeopleConfirm+", their keys and sessions?") +
			ui.Dim.Render("  [y] remove, any other key keeps them") + "\n")
	}
	if m.PeopleSessions != nil {
		b.WriteString("\n" + ui.Title.Render("▌SESSIONS") + "\n\n")
		for _, s := range m.PeopleSessions {
			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				ui.Value.Render(fmt.Sprintf("%-12s", s.User)),
				ui.Label.Render(fmt.Sprintf("%-10s", s.Panel)),
				ui.Dim.Render("since "+s.Since.Local().Format("2006-01-02 15:04"))))
		}
	}
	return b.String()
}

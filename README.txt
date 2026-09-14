███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝


  ░▒▓█ _monoview_ █▓▒░
  The **TUI monitor** for MONOLITH — keeping watch over your system.

  ───────────────────────────────────────────────────────────────
  ▓ OVERVIEW
  **monoview** is a MONOLITH **client** written in **Go** — a **terminal UI (TUI)**.
  ▪ Connects to **concentrator** over `wss://` with mTLS, the hub checked by the bubble CA — and nothing else
  ▪ Controls **VERTEX**, manages **achtung** timers and alarms, and monitors nodes from the terminal
  ▪ UI stack: Bubble Tea, Lipgloss, Gorilla WebSocket, pflag

  ───────────────────────────────────────────────────────────────
  ▓ ARCHITECTURE
  ▪ **RUNTIME**: Go 1.25+ (see `go.mod`)
  ▪ **TRANSPORT**: monolink v2 over `wss://` with mTLS (`github.com/MrZloHex/monolink` v0.3.1)
  ▪ **NODE ID**: `MONOVIEW` (in code)

  ───────────────────────────────────────────────────────────────
  ▓ FEATURES
  ▪ Six sheets: Calendar, Diary, Home, System, People, Synapse
  ▪ Messages with the bubble's other people through **SYNAPSE**, unread counted on the tab
  ▪ **VERTEX** device control (lamps, LEDs, brightness), through the properties uart2ws registers for it — `LAMP.STATE`, `LED.STATE`, `LED.MODE`, `LED.BRIGHT`, `BUZZ.STATE` — set with SET and followed by their PUBs
  ▪ **ACHTUNG** jobs — timers, alarms, repeating intervals and daily wall-clock jobs (create, list, delete; realtime countdown)
  ▪ Fire alert when a timer or alarm fires (turn off buzzer)
  ▪ Node status (ping, uptime) and recent hub message log
  ▪ Signing in through **MARSHAL**: what this panel sends is then `MONOVIEW.<person>`, and only what their grants allow

  ───────────────────────────────────────────────────────────────
  ▓ SHEETS
  ▪ **[1] CALENDAR** — Events, weekly schedule and deadlines, live from **GOVERNOR**
  ▪ **[2] DIARY** — Entries with mood (sample data)
  ▪ **[3] HOME** — **VERTEX** devices (toggle, cycle, value) and **ACHTUNG** timers and alarms
  ▪ **[4] SYSTEM** — Node panels (**VERTEX**, **ACHTUNG**, **GOVERNOR**, **UKAZ**, **MARSHAL**, **SYNAPSE**), ping, uptime, recent concentrator messages
  ▪ **[5] PEOPLE** — Who is signed in here; for whoever may, the people of the bubble, their grants and sessions
  ▪ **[6] SYNAPSE** — Conversations with the other people of the bubble: unread, read receipts (✓), live as messages arrive. Needs someone signed in; a long text goes as several messages

  ───────────────────────────────────────────────────────────────
  ▓ CONTROLS
  Global:
    [1]–[6] or [Tab] / [Shift+Tab]   Switch sheet
    [Q] / [Ctrl+C]                    Quit

  Calendar:  [←/h] [→/l]   Prev/next day
  Diary:     [↑/k] [↓/j]   Prev/next entry
  Home:      [Tab]         Focus devices ↔ timers (ACHTUNG)
             Devices:     [↑/k ↓/j] select  [Enter] toggle  [←/h →/l] adjust
             Jobs:        [↑/k ↓/j] job  [t] timer  [a] alarm  [e] every  [D] daily  [d] delete
  System:    [↑/k ↓/j] or [←/h →/l] select node  [Enter] ping
  People:    [s] sign in  [e] first person (with the code MARSHAL printed)  [i] invitation  [o] sign out
             [↑/k ↓/j] person  [n] invite  [g] grant  [x] revoke  [K] remove a key  [D] remove  [r] refresh
  Synapse:   [↑/k ↓/j] person  [Enter] open  [i] write, [Enter] send, [Esc] stop  [u] earlier  [r] refresh

  It opens on the sign-in form and sends nothing until someone signs in: the
  hub holds this panel to the ticket marshal signs for its person, renewed
  every four minutes (SECURITY.txt §5). Signed in, it refuses — and logs
  once — anything the person's grants do not cover.

  It signs in with a key of its own: Ed25519, in `monoview.key`, sealed
  with the person's passphrase (argon2id, AES-256-GCM). Enrolment or an
  invitation makes it; signing in opens it, signs MARSHAL's challenge, and
  forgets it. The file alone, copied, signs nobody in. The session is kept
  nowhere: a restart asks for the passphrase again.

  A change to people, keys or grants needs a sign-in within the last five
  minutes — MARSHAL takes one only from a session that fresh — so [n], [K],
  [g], [x] and [D] open the sign-in form first when the last is older.

  Fire alert popup:  [Enter] / [Space]  Turn off buzzer and close

  ───────────────────────────────────────────────────────────────
  ▓ REQUIREMENTS
  ▪ Go 1.25+ (see `go.mod`)
  ▪ A terminal with alternate-screen support (Bubble Tea)

  ───────────────────────────────────────────────────────────────
  ▓ BUILD & RUN
  **Build**
  ```sh
  go build -o bin/monoview ./cmd/monoview
  ```

  **Run**
  ```sh
  ./bin/monoview
  ```
  Default hub URL is `wss://127.0.0.1:8443` unless overridden — see **CONFIGURATION**. If the hub is unreachable, the app still starts; the hub indicator shows offline until connected.

  **Example** (certificates + log path)
  ```sh
  ./bin/monoview --tls-cert monoview.pem --tls-key monoview.key.pem --tls-ca bubble-ca.pem \
    --log-path /tmp/monoview.log
  ```
  It does not start without a `wss://` URL and all three TLS files.

  ───────────────────────────────────────────────────────────────
  ▓ CONFIGURATION
  Dotenv is loaded before flags: path is `MONO_ENV_FILE`, or `--env-file` from argv, or `.env`. A missing file is ignored; parse errors exit with an error message.

  **Environment**
  ▪ `MONOVIEW_URL` — WebSocket URL (default `wss://127.0.0.1:8443`)
  ▪ `MONOVIEW_LOG` — log file path (default `monoview.log`)
  ▪ `MONOVIEW_TLS_CERT` — client certificate PEM (mTLS)
  ▪ `MONOVIEW_TLS_KEY` — client private key PEM (mTLS)
  ▪ `MONOVIEW_TLS_CA` — the bubble CA's PEM, which vouches for the hub (required)
  ▪ `MONOVIEW_TLS_SERVER_NAME` — TLS ServerName (SNI); e.g. when dialing an IP
  ▪ `MONOVIEW_KEY` — this panel's sealed key (default `monoview.key`)
  ▪ `MONO_ENV_FILE` — path to dotenv file instead of `.env`

  **Flags** (see `./bin/monoview --help`)
  ▪ `-u`, `--url` — hub URL (`MONOVIEW_URL`)
  ▪ `--tls-cert`, `--tls-key` — client mTLS (`MONOVIEW_TLS_*`)
  ▪ `--tls-ca` — the bubble CA (`MONOVIEW_TLS_CA`)
  ▪ `--tls-server-name` — SNI (`MONOVIEW_TLS_SERVER_NAME`)
  ▪ `--log-path` — log file (`MONOVIEW_LOG`)
  ▪ `--key` — the sealed key file (`MONOVIEW_KEY`)
  ▪ `--env-file` — dotenv path (early parse)

  **Example** (environment overrides)
  ```sh
  MONOVIEW_URL=wss://hub.example:8443 ./bin/monoview
  ```

  ───────────────────────────────────────────────────────────────
  ▓ PROTOCOL
  monolink v2 (`2:<id>:<from>:<to>:<verb>:<noun>[:<arg>...]`, SPEC Part III), sent as `MONOVIEW.<person>`. Shared client and parsing live in `../monolink`; UI wiring under `internal/app`. GOVERNOR's events and ACHTUNG's jobs come a frame's worth at a time; monoview asks for every page.

  ───────────────────────────────────────────────────────────────
  ▓ ACHTUNG (HOME SHEET)
  On the Home sheet, focus the **ACHTUNG** panel ([Tab]) then:
  ▪ **[t] Timer** — Duration (presets or [c] custom), then name (or Enter for auto). Time-to-fire updates every second.
  ▪ **[e] Every** — Repeating interval (e.g. `45m`; a minute or more), then name. Repeats from now.
  ▪ **[D] Daily** — A local wall-clock `HH:MM`, then name. Fires every day at that time, stays put across DST, and survives an **achtung** restart. This is what drives the morning agenda printout.
  ▪ **[a] Alarm** — One-shot; pick when ([1] today, [2] tomorrow, [c] custom). Custom: `HH:MM`; if that time passed today, alarm is set for tomorrow.
  ▪ **[d]** / **[Enter]** on a job — Stop or delete it.
  The job list syncs with **achtung** about every minute.

  ───────────────────────────────────────────────────────────────
  ▓ FINAL WORDS
  This is not just a monitor. This is **monoview** — the eyes of MONOLITH.

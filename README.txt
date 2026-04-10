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
  ▪ Connects to **concentrator** over WebSocket (`ws://` or `wss://` with optional mTLS)
  ▪ Controls **VERTEX**, manages **achtung** timers and alarms, and monitors nodes from the terminal
  ▪ UI stack: Bubble Tea, Lipgloss, Gorilla WebSocket, pflag

  ───────────────────────────────────────────────────────────────
  ▓ ARCHITECTURE
  ▪ **RUNTIME**: Go 1.25+ (see `go.mod`)
  ▪ **TRANSPORT**: WebSocket client (`github.com/MrZloHex/monolink`); optional **mTLS** (`wss://`)
  ▪ **NODE ID**: `MONOVIEW` (in code)

  ───────────────────────────────────────────────────────────────
  ▓ FEATURES
  ▪ Four sheets: Calendar, Diary, Home, System
  ▪ **VERTEX** device control (lamps, LEDs, brightness)
  ▪ **ACHTUNG** timers and alarms (create, list, delete; realtime countdown)
  ▪ Fire alert when a timer or alarm fires (turn off buzzer)
  ▪ Node status (ping, uptime) and recent hub message log

  ───────────────────────────────────────────────────────────────
  ▓ SHEETS
  ▪ **[1] CALENDAR** — Events and weekly schedule (sample data)
  ▪ **[2] DIARY** — Entries with mood (sample data)
  ▪ **[3] HOME** — **VERTEX** devices (toggle, cycle, value) and **ACHTUNG** timers and alarms
  ▪ **[4] SYSTEM** — Node panels (**VERTEX**, **ACHTUNG**), ping, uptime, recent concentrator messages

  ───────────────────────────────────────────────────────────────
  ▓ CONTROLS
  Global:
    [1]–[4] or [Tab] / [Shift+Tab]   Switch sheet
    [Q] / [Ctrl+C]                    Quit

  Calendar:  [←/h] [→/l]   Prev/next day
  Diary:     [↑/k] [↓/j]   Prev/next entry
  Home:      [Tab]         Focus devices ↔ timers (ACHTUNG)
             Devices:     [↑/k ↓/j] select  [Enter] toggle  [←/h →/l] adjust
             Timers:      [↑/k ↓/j] job  [t] timer  [a] alarm  [d] delete
  System:    [↑/k ↓/j] or [←/h →/l] select node  [Enter] ping

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

  **Example** (plain WebSocket + log path)
  ```sh
  ./bin/monoview -u ws://localhost:8092 --log-path /tmp/monoview.log
  ```

  ───────────────────────────────────────────────────────────────
  ▓ CONFIGURATION
  Dotenv is loaded before flags: path is `MONO_ENV_FILE`, or `--env-file` from argv, or `.env`. A missing file is ignored; parse errors exit with an error message.

  **Environment**
  ▪ `MONOVIEW_URL` — WebSocket URL (default `wss://127.0.0.1:8443`)
  ▪ `MONOVIEW_LOG` — log file path (default `monoview.log`)
  ▪ `MONOVIEW_TLS_CERT` — client certificate PEM (mTLS)
  ▪ `MONOVIEW_TLS_KEY` — client private key PEM (mTLS)
  ▪ `MONOVIEW_TLS_CA` — optional CA PEM to verify the server
  ▪ `MONOVIEW_TLS_SERVER_NAME` — TLS ServerName (SNI); e.g. when dialing an IP
  ▪ `MONO_ENV_FILE` — path to dotenv file instead of `.env`

  **Flags** (see `./bin/monoview --help`)
  ▪ `-u`, `--url` — hub URL (`MONOVIEW_URL`)
  ▪ `--tls-cert`, `--tls-key` — client mTLS (`MONOVIEW_TLS_*`)
  ▪ `--tls-ca` — optional server CA (`MONOVIEW_TLS_CA`)
  ▪ `--tls-server-name` — SNI (`MONOVIEW_TLS_SERVER_NAME`)
  ▪ `--log-path` — log file (`MONOVIEW_LOG`)
  ▪ `--env-file` — dotenv path (early parse)

  **Example** (environment overrides)
  ```sh
  MONOVIEW_URL=wss://hub.example:8443 ./bin/monoview
  ```

  ───────────────────────────────────────────────────────────────
  ▓ PROTOCOL
  Wire format: `TO:VERB:NOUN[:ARGS]:FROM` (DSKY-style). Shared client and parsing live in `../monolink`; UI wiring under `internal/app`.

  ───────────────────────────────────────────────────────────────
  ▓ ACHTUNG (HOME SHEET)
  On the Home sheet, focus the **ACHTUNG** panel ([Tab]) then:
  ▪ **[t] Timer** — Duration (presets or [c] custom), then name (or Enter for auto). Time-to-fire updates every second.
  ▪ **[a] Alarm** — One-shot; pick when ([1] today, [2] tomorrow, [c] custom). Custom: `HH:MM`; if that time passed today, alarm is set for tomorrow.
  ▪ **[d]** / **[Enter]** on a job — Stop or delete it.
  The job list syncs with **achtung** about every minute.

  ───────────────────────────────────────────────────────────────
  ▓ FINAL WORDS
  This is not just a monitor. This is **monoview** — the eyes of MONOLITH.

███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝

  ░▒▓█ _MonoView_ █▓▒░
  The **TUI monitor** for MONOLITH-keeping watch over your system.

  ───────────────────────────────────────────────────────────────
  ▓ OVERVIEW
  **_MonoView_** is a **terminal UI (TUI)** for the **MONOLITH** ecosystem.
  It connects to a **concentrator** hub over WebSocket and provides
  device control, timers & alarms (ACHTUNG), and node monitoring.

  **Tech:** Go, Bubble Tea (TUI), Lipgloss (styling), Gorilla WebSocket, pflag.
  **Wire format:** TO:VERB:NOUN[:ARGS]:FROM (DSKY-style).

  **Features at a glance:**
  ▪ Four sheets: Calendar, Diary, Home, System
  ▪ VERTEX device control (lamps, LEDs, brightness)
  ▪ ACHTUNG timers & alarms (create, list, delete; realtime countdown)
  ▪ Fire alert popup when a timer/alarm fires (turn off buzzer)
  ▪ Node status (ping, uptime) and recent message log

  ───────────────────────────────────────────────────────────────
  ▓ SHEETS
  ▪ **[1] CALENDAR** – Events and weekly schedule (sample data)
  ▪ **[2] DIARY**    – Entries with mood (sample data)
  ▪ **[3] HOME**     – VERTEX devices (toggle, cycle, value) and
                      ACHTUNG timers & alarms (new timer/alarm, delete)
  ▪ **[4] SYSTEM**   – Node panels (VERTEX, ACHTUNG), ping, uptime,
                      recent concentrator messages

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
  ▓ BUILD & RUN
  Requirements: Go version as in go.mod (currently 1.25.x).

  Build:
    go build -o monoview ./cmd/monoview

  Run (defaults: node name MONOVIEW in code, url wss://127.0.0.1:8443):
    ./monoview

  Optional .env: loaded from .env unless MONO_ENV_FILE is set or you pass
    --env-file /path/to/.env
  (Missing file is ignored; parse errors exit with a message.)

  Environment (override defaults; same keys work in .env):
    MONOVIEW_URL           WebSocket URL (default: wss://127.0.0.1:8443)
    MONOVIEW_LOG           Log file path (default: monoview.log)
    MONOVIEW_TLS_CERT      Client cert PEM for mTLS (wss)
    MONOVIEW_TLS_KEY       Client private key PEM for mTLS
    MONOVIEW_TLS_CA        Optional CA PEM to verify the server
    MONOVIEW_TLS_SERVER_NAME  TLS ServerName (SNI); e.g. when dialing an IP
    MONO_ENV_FILE          Path to dotenv file instead of .env

  CLI (see ./monoview --help):
    -u, --url              Hub URL
    --tls-cert, --tls-key, --tls-ca, --tls-server-name
    --log-path             Log file
    --env-file             Dotenv path (before full parse; see above)

  Example:
    MONOVIEW_URL=wss://hub.example:8443 ./monoview
    ./monoview -u ws://localhost:8092 --log-path /tmp/monoview.log

  If the hub is unreachable, the app still starts; the hub indicator
  shows offline and features wait for connection.

  ───────────────────────────────────────────────────────────────
  ▓ ACHTUNG (timers & alarms)
  On the Home sheet, focus the ACHTUNG panel ([Tab]) then:
  ▪ **[t] Timer** – Pick duration (presets or [c] custom), then name
                   (or Enter for auto). Time-till updates every second.
  ▪ **[a] Alarm** – Pick type [1] One-shot, when ([1] today 20:00,
                   [2] tomorrow 08:00, [c] custom time), then name.
                   Custom time: enter HH:MM; if that time passed today,
                   alarm is set for tomorrow.
  ▪ **[d] / [Enter]** on a job – Stop/delete it.
  List syncs with ACHTUNG every minute. Wire format is TO:VERB:NOUN[:ARGS]:FROM
  (DSKY-style); see pkg/concentrator and internal/app for message handling.

  ───────────────────────────────────────────────────────────────
  ▓ FINAL WORDS
  This is not just a monitor.
  This is **_MonoView_**-the eyes of MONOLITH.

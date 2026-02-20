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

  **Tech:** Go, Bubble Tea (TUI), Lipgloss (styling), Gorilla WebSocket.
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
  Requirements: Go 1.21+ (see go.mod).

  Build:
    go build -o monoview ./cmd/monoview

  Run (defaults: node=MONOVIEW, url=ws://192.168.0.69:8092):
    ./monoview

  Environment:
    MONO_NODE   Node name (default: MONOVIEW)
    MONO_URL    Concentrator WebSocket URL (default: ws://192.168.0.69:8092)
    MONO_LOG    Log file path (default: monoview.log)

  Example:
    MONO_NODE=MONOVIEW MONO_URL=ws://localhost:8092 ./monoview

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
  List syncs with ACHTUNG every minute. See README_achtung.txt for
  protocol details.

  ───────────────────────────────────────────────────────────────
  ▓ FINAL WORDS
  This is not just a monitor.
  This is **_MonoView_**-the eyes of MONOLITH.

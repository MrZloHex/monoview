███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝

  ░▒▓█ _MonoView_ █▓▒░
  The **TUI monitor** for MONOLITH—keeping watch over your system.

  ───────────────────────────────────────────────────────────────
  ▓ OVERVIEW
  **_MonoView_** is a **terminal-based user interface (TUI) monitor**
  for **MONOLITH**, providing real-time system diagnostics,
  device status, and network monitoring.
  Running on **HAL9000**, it pulls data from **TMA** and
  displays live updates for **_obelisk_**, **_vertex_**, and other nodes.

  **Features at a glance:**
  ▪ Live telemetry from MONOLITH devices
  ▪ System stats (CPU, memory, network)
  ▪ USART/WebSocket data streams
  ▪ Minimalist, fast, and low-resource

  ───────────────────────────────────────────────────────────────
  ▓ HARDWARE / ENVIRONMENT
  ▪ **HOST**: HAL9000 (Old Laptop)
  ▪ **LANGUAGE**: Pure C
  ▪ **UI LIBRARY**: ncurses
  ▪ **DATA SOURCES**: TMA (WebSocket), /dev/serial (USART)

  ───────────────────────────────────────────────────────────────
  ▓ FEATURES
  ▪ **TUI Interface** – Real-time monitoring in the terminal
  ▪ **Device Status Dashboard** – Watch _obelisk_, _vertex_, and more
  ▪ **System Resource Tracking** – CPU, RAM, Network, Disk
  ▪ **WebSocket & USART** – Live feeds from MONOLITH nodes
  ▪ **Minimalist & Efficient** – No bloat, just raw data

  ───────────────────────────────────────────────────────────────
  ▓ BUILD & RUN
  _MonoView_ is written in **pure C**, using **ncurses** for TUI rendering.

  **Dependencies:**
  - **GCC** (or Clang)
  - **ncurses**
  - **libwebsockets**

  **Build Instructions:**
  mkdir build && cd build
  cmake ..
  make

  **Run the Monitor:**
  ./monoview

  ───────────────────────────────────────────────────────────────
  ▓ CONTROLS
  _MonoView_ is fully keyboard-driven.
  Use the following keys to navigate and interact:

  [↑] / [↓]    Scroll through device list
  [Enter]      Show device details
  [R]          Refresh data
  [Q]          Quit

  ───────────────────────────────────────────────────────────────
  ▓ FINAL WORDS
  This is not just a monitor.
  This is **_MonoView_**—the eyes of MONOLITH.

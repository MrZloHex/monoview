#!/usr/bin/env python3
import argparse
import csv
import json
import os
import shutil
import sys
import textwrap
from dataclasses import dataclass
from datetime import date, datetime, time, timedelta
from typing import Dict, List, Optional, Tuple

# ----- ANSI + wide char width helpers -----
import re, unicodedata
ANSI_RE = re.compile(r'\x1b\[[0-9;?]*[ -/]*[@-~]')

def strip_ansi(s: str) -> str:
    return ANSI_RE.sub('', s)

def _char_width(ch: str) -> int:
    # Zero width for combining marks & format chars (ZWJ etc.)
    if unicodedata.combining(ch) != 0:
        return 0
    cat = unicodedata.category(ch)
    if cat == 'Cf':
        return 0
    eaw = unicodedata.east_asian_width(ch)
    if eaw in ('F','W'):
        return 2
    # Treat 'A' (ambiguous) as 1 in most terminals
    return 1

def wcswidth(s: str) -> int:
    s = strip_ansi(s)
    return sum(_char_width(ch) for ch in s)

def truncate_ansi(s: str, maxw: int) -> str:
    """Cut string to display width maxw without breaking ANSI, add RESET to avoid bleed."""
    out = []
    width = 0
    i = 0
    while i < len(s):
        m = ANSI_RE.match(s, i)
        if m:
            out.append(m.group(0))
            i = m.end()
            continue
        ch = s[i]
        w = _char_width(ch)
        if width + w > maxw:
            break
        out.append(ch)
        width += w
        i += 1
    # ensure styles reset at end so borders don't inherit
    return ''.join(out) + RESET

def wrap_ansi(s: str, width: int):
    """Simple greedy wrapper by display width, preserving ANSI."""
    lines = []
    cur = []
    curw = 0
    i = 0
    while i < len(s):
        m = ANSI_RE.match(s, i)
        if m:
            cur.append(m.group(0))
            i = m.end()
            continue
        ch = s[i]
        if ch == '\n':
            lines.append(''.join(cur) + RESET)
            cur, curw = [], 0
            i += 1
            continue
        w = _char_width(ch)
        if curw + w > width:
            lines.append(''.join(cur) + RESET)
            cur, curw = [], 0
        else:
            cur.append(ch)
            curw += w
            i += 1
    if cur or not lines:
        lines.append(''.join(cur) + RESET)
    return lines

def pad_ansi(s: str, width: int) -> str:
    vis = wcswidth(s)
    if vis < width:
        return s + ' '*(width - vis) + RESET
    return truncate_ansi(s, width)


try:
    from zoneinfo import ZoneInfo
except Exception:
    ZoneInfo = None

RESET = "\x1b[0m"
BOLD = "\x1b[1m"
DIM = "\x1b[2m"

def hex_to_rgb(hexstr: str) -> Tuple[int,int,int]:
    s = hexstr.strip().lstrip('#')
    if len(s) == 3:
        s = ''.join(c*2 for c in s)
    if len(s) != 6:
        raise ValueError(f"Bad hex color: {hexstr}")
    return int(s[0:2],16), int(s[2:4],16), int(s[4:6],16)

def rgb_fg(r,g,b): return f"\x1b[38;2;{r};{g};{b}m"
def rgb_bg(r,g,b): return f"\x1b[48;2;{r};{g};{b}m"

def contrast_text(hexstr: str) -> str:
    r,g,b = hex_to_rgb(hexstr)
    yiq = (r*299 + g*587 + b*114) / 1000
    return "#000000" if yiq > 128 else "#FFFFFF"

@dataclass
class Event:
    d: date
    start: str
    end: str
    title: str
    location: str
    tags: List[str]
# ---------------- Weekly support ----------------

WEEKDAY_MAP = {
    "mon": 0, "monday": 0,
    "tue": 1, "tues": 1, "tuesday": 1,
    "wed": 2, "wednesday": 2,
    "thu": 3, "thur": 3, "thurs": 3, "thursday": 3,
    "fri": 4, "friday": 4,
    "sat": 5, "saturday": 5,
    "sun": 6, "sunday": 6,
}

@dataclass
class WeeklyRow:
    weekday: int  # 0=Mon..6=Sun
    start: str
    end: str
    title: str
    location: str
    tags: List[str]

def parse_weekday_value(val: str) -> int:
    s = (val or "").strip().lower()
    if s in WEEKDAY_MAP:
        return WEEKDAY_MAP[s]
    # numeric support: 1=Mon..7=Sun OR 0=Mon..6=Sun
    try:
        n = int(s)
        if 0 <= n <= 6:
            return n
        if 1 <= n <= 7:
            return (n - 1) % 7
    except Exception:
        pass
    raise SystemExit(f"Bad weekday value: {val!r}. Use Mon..Sun or 0..6 or 1..7.")

def parse_csv_weekly(path: str) -> List[WeeklyRow]:
    req = {"weekday","start","title"}
    rows: List[WeeklyRow] = []
    with open(path, newline='', encoding='utf-8') as f:
        r = csv.DictReader(f)
        hdr = { (h or "").strip().lower() for h in (r.fieldnames or []) }
        missing = req - hdr
        if missing:
            raise SystemExit(f"Missing required columns for weekly mode: {', '.join(sorted(missing))}")
        for row in r:
            if not any((row.get(k) or "").strip() for k in row):
                continue
            wd = parse_weekday_value(row.get("weekday",""))
            start = (row.get("start") or "").strip()
            end = (row.get("end") or "").strip()
            title = (row.get("title") or "").strip()
            location = (row.get("location") or "").strip()
            tags_str = (row.get("tags") or "").strip()
            tags = [t.strip() for t in tags_str.replace(",", ";").split(";") if t.strip()]
            rows.append(WeeklyRow(wd, start, end, title, location, tags))
    rows.sort(key=lambda e: (e.weekday, e.start, e.end, e.title))
    return rows

def expand_weekly_to_events(rows: List[WeeklyRow], ref_date) -> List[Event]:
    # ref_date is a date in the target week; compute Monday..Sunday
    monday = ref_date - timedelta(days=ref_date.weekday())
    events: List[Event] = []
    for r in rows:
        d = monday + timedelta(days=r.weekday)
        events.append(Event(d, r.start, r.end, r.title, r.location, r.tags))
    events.sort(key=lambda e: (e.d, e.start, e.end, e.title))
    return events


def parse_csv(path: str) -> List[Event]:
    req = {"date","start","title"}
    events: List[Event] = []
    with open(path, newline='', encoding='utf-8') as f:
        r = csv.DictReader(f)
        hdr = { (h or "").strip().lower() for h in (r.fieldnames or []) }
        missing = req - hdr
        if missing:
            raise SystemExit(f"Missing required columns: {', '.join(sorted(missing))}")
        for row in r:
            if not any((row.get(k) or "").strip() for k in row):
                continue
            d = datetime.strptime((row.get("date") or "").strip(), "%Y-%m-%d").date()
            start = (row.get("start") or "").strip()
            end = (row.get("end") or "").strip()
            title = (row.get("title") or "").strip()
            location = (row.get("location") or "").strip()
            tags_str = (row.get("tags") or "").strip()
            tags = [t.strip() for t in tags_str.replace(",", ";").split(";") if t.strip()]
            events.append(Event(d, start, end, title, location, tags))
    events.sort(key=lambda e: (e.d, e.start, e.end, e.title))
    return events

def load_palette(path: Optional[str]) -> Dict[str,str]:
    if not path:
        return {}
    with open(path, encoding="utf-8") as f:
        data = json.load(f)
    return {str(k).strip(): str(v).strip() for k,v in data.items()}

# ---------------- Theme ----------------

def theme_colors(name: str) -> Dict[str, Tuple[int,int,int]]:
    name = (name or "dark").lower()
    if name in ("gruvbox", "gruvbox-dark", "gruvboxd"):
        # Gruvbox dark palette
        return {
            "fg": hex_to_rgb("#ebdbb2"),
            "muted": hex_to_rgb("#928374"),
            "border": hex_to_rgb("#3c3836"),
            "rule": hex_to_rgb("#504945"),
            "highlight": hex_to_rgb("#fabd2f"),  # yellow
            "dim": hex_to_rgb("#7c6f64"),
            "stripe_fallback": hex_to_rgb("#928374"),
        }
    else:
        # default dark
        return {
            "fg": hex_to_rgb("#e6edf3"),
            "muted": hex_to_rgb("#9aa7b3"),
            "border": hex_to_rgb("#374151"),
            "rule": hex_to_rgb("#4b5563"),
            "highlight": hex_to_rgb("#f59e0b"),  # amber
            "dim": hex_to_rgb("#9aa7b3"),
            "stripe_fallback": hex_to_rgb("#9aa7b3"),
        }

def colorize(s: str, rgb: Tuple[int,int,int]) -> str:
    return rgb_fg(*rgb) + s + RESET

def badge(text: str, color_hex: Optional[str]) -> str:
    if not color_hex:
        return f"[{text}]"
    r,g,b = hex_to_rgb(color_hex)
    tc = contrast_text(color_hex)
    tr,tg,tb = hex_to_rgb(tc)
    return f"{rgb_bg(r,g,b)}{rgb_fg(tr,tg,tb)}[{text}]{RESET}"

def first_tag_color(tags: List[str], palette: Dict[str,str]) -> Optional[str]:
    for t in tags:
        if t in palette:
            return palette[t]
    return None

def parse_hhmm(s: str) -> Optional[time]:
    s = (s or "").strip()
    if not s:
        return None
    try:
        hh, mm = s.split(":")
        return time(int(hh), int(mm))
    except Exception:
        return None

def status_for_event(e: Event, now_dt: datetime, tz: Optional[str]) -> str:
    """Return 'past' | 'current' | 'future' relative to now_dt (which is already tz-aware)."""
    if e.d != now_dt.date():
        return "future" if e.d > now_dt.date() else "past"
    st = parse_hhmm(e.start) or time(0,0)
    en = parse_hhmm(e.end)
    if en is None:
        # assume 1h if no end
        en = (datetime.combine(e.d, st) + timedelta(hours=1)).time()
    start_dt = datetime.combine(e.d, st, tzinfo=now_dt.tzinfo)
    end_dt = datetime.combine(e.d, en, tzinfo=now_dt.tzinfo)
    if start_dt <= now_dt <= end_dt:
        return "current"
    if now_dt > end_dt:
        return "past"
    return "future"



def render_event_card(e: Event, palette: Dict[str,str], width: int, theme: Dict[str,Tuple[int,int,int]], status: str, dim_past: bool) -> str:
    maxw = max(40, min(width, 100))
    inner = maxw - 2
    pad_l = 1
    content_w = inner - pad_l -  3  # 2 cells for stripe + 1 space separator

    primary_hex = first_tag_color(e.tags, palette) or "#%02x%02x%02x" % theme["stripe_fallback"]
    pr,pg,pb = hex_to_rgb(primary_hex)
    stripe = f"{rgb_bg(pr,pg,pb)}  {RESET}"

    border_rgb = theme["border"]
    if status == "current":
        border_rgb = theme["highlight"]
    def bcol(s: str) -> str:
        return colorize(s, border_rgb)

    top = bcol("┌" + "─"*inner + "┐")
    bot = bcol("└" + "─"*inner + "┘")

    def line(txt=""):
        padded = pad_ansi(txt, content_w)
        return bcol("│") + " "*pad_l + stripe + " " + padded + bcol("│")

    times = e.start + (f"–{e.end}" if e.end else "")
    title = f"⏰ {times}  {BOLD}{e.title}{RESET}"
    location = f"{DIM}📍 {e.location}{RESET}" if e.location else ""
    badges = " ".join(badge(t, palette.get(t)) for t in e.tags)
    if status == "current":
        now_badge = badge("NOW", "#fabd2f")
        badges = (now_badge + "  " + badges) if badges else now_badge

    wrap_w = content_w
    title_lines = wrap_ansi(title, wrap_w)
    meta_parts = []
    if badges: meta_parts.append(badges)
    if location: meta_parts.append(location)
    meta_line = "  ".join(meta_parts)
    meta_lines = wrap_ansi(meta_line, wrap_w) if meta_line else []

    pre = DIM if dim_past else ""
    post = RESET if dim_past else ""

    out = [pre + top]
    for tl in title_lines:
        out.append(pre + line(tl))
    for ml in meta_lines:
        out.append(pre + line(ml))
    out.append(pre + bot + post)
    return "\n".join(out)

def print_agenda(events: List[Event], palette: Dict[str,str], theme_name: str, now_dt: datetime):
    if not events:
        print("No events.")
        return
    theme = theme_colors(theme_name)
    cols = shutil.get_terminal_size((90, 24)).columns
    rule = colorize("─"*min(cols-2, 60), theme["rule"])
    current_day = None
    today = now_dt.date()
    for e in events:
        if e.d != current_day:
            current_day = e.d
            hdr = f"📅 {current_day.strftime('%A, %Y-%m-%d')}"
            print("\n" + colorize(BOLD + hdr + RESET, theme["fg"]))
            print(rule)
        status = status_for_event(e, now_dt, None)
        dim_past = (status == "past" and e.d == today)
        print(render_event_card(e, palette, cols, theme, status, dim_past))
        print()

def render_compact_line(e: Event, palette: Dict[str,str], theme: Dict[str,Tuple[int,int,int]], status: str, width: int) -> str:
    """One-line event: colored square, time, title, tags, location. Dims past; adds [NOW] badge."""
    cols = max(50, min(width, 140))
    sq_hex = first_tag_color(e.tags, palette) or "#%02x%02x%02x" % theme["stripe_fallback"]
    r,g,b = hex_to_rgb(sq_hex)
    square = rgb_fg(r,g,b) + "■" + RESET  # colored indicator
    times = e.start + (f"-{e.end}" if e.end else "")
    title = f"{BOLD}{e.title}{RESET}"
    if status == "current":
        title = rgb_fg(0xfe, 0x80, 0x19) + "-> " + RESET + title
    pad_title = pad_ansi(title, 26)
    badges = " ".join(badge(t, palette.get(t)) for t in e.tags)
    loc = f"{DIM}{e.location}{RESET}" if e.location else ""
    pad_loc = pad_ansi(loc, 7)
    # Compose pieces and then truncate to terminal width
    parts = [square, times]
    if loc: parts.append(pad_loc)
    parts.append(pad_title)
    if badges: parts.append(badges)
    line = "  ".join(parts)
    return truncate_ansi(line, cols)

def print_agenda_compact(events: List[Event], palette: Dict[str,str], theme_name: str, now_dt: datetime):
    if not events:
        print("No events.")
        return
    theme = theme_colors(theme_name)
    cols = shutil.get_terminal_size((90, 24)).columns
    rule = colorize("─"*min(cols-2, 60), theme["rule"])
    current_day = None
    today = now_dt.date()
    for e in events:
        if e.d != current_day:
            current_day = e.d
            hdr = f"📅 {current_day.strftime('%A, %Y-%m-%d')}"
            print("\n" + colorize(BOLD + hdr + RESET, theme["fg"]))
            print(rule)
        status = status_for_event(e, now_dt, None)
        dim_past = (status == "past" and e.d == today)
        line = render_compact_line(e, palette, theme, status, cols)
        if dim_past:
            line = DIM + line + RESET
        print("  " + line)  # small indent for readability

    print("\n")

def main():
    ap = argparse.ArgumentParser(description="Colorful university schedule (CLI only)")
    ap.add_argument("csv", help="Schedule CSV with columns: date,start,end,title,location,tags")
    ap.add_argument("--tags", help="JSON mapping of tag -> hex color (#RRGGBB)", default=None)
    ap.add_argument("--today", action="store_true", help="Show only today's entries")
    ap.add_argument("--this-week", action="store_true", help="Show only this week's entries (Mon–Sun)")
    ap.add_argument("--theme", default="dark", choices=["dark","gruvbox","gruvbox-dark"], help="Theme for borders/headings")
    ap.add_argument("--compact", action="store_true", help="Force compact list view (auto when multiple days)")
    ap.add_argument("--tz", default="", help="IANA timezone (e.g. Europe/Chisinau). Uses system TZ if omitted.")
    ap.add_argument("--now", default="", help="Override current time for testing, format YYYY-MM-DDTHH:MM")
    args = ap.parse_args()

    # Detect mode: date schedule vs weekly schedule
    # Look at header to decide
    with open(args.csv, newline='', encoding='utf-8') as _f:
        _r = csv.reader(_f)
        _hdr = [h.strip().lower() for h in next(_r, [])]
    weekly_mode = ('weekday' in _hdr) and ('date' not in _hdr)
    if weekly_mode:
        weekly_rows = parse_csv_weekly(args.csv)
        # Default to this week if no filter provided
        # Expand relative to now_dt date (computed below)
        pass
    else:
        events = parse_csv(args.csv)
    palette = load_palette(args.tags)

    # Determine 'now' with timezone (needed before expanding weekly)
    if args.now:
        try:
            naive = datetime.strptime(args.now, "%Y-%m-%dT%H:%M")
        except Exception as e:
            raise SystemExit(f"--now format must be YYYY-MM-DDTHH:MM, got {args.now}")
        if args.tz and ZoneInfo:
            now_dt = naive.replace(tzinfo=ZoneInfo(args.tz))
        else:
            now_dt = naive
    else:
        if args.tz and ZoneInfo:
            now_dt = datetime.now(ZoneInfo(args.tz))
        else:
            now_dt = datetime.now()

    # Expand weekly if needed
    if 'weekly_mode' in locals() and weekly_mode:
        events = expand_weekly_to_events(weekly_rows, now_dt.date())

    # Defaults; swap to Gruvbox palette if chosen
    if args.theme.startswith("gruvbox"):
        defaults = {
            "Lecture":"#83a598",  # blue
            "Lab":"#d3869b",      # purple
            "Seminar":"#8ec07c",  # aqua
            "Study":"#b8bb26",    # green
            "Exam":"#fb4934",     # red
            "EE":"#fe8019",       # orange
            "Math":"#b16286",     # purple (alt)
            "CS":"#8ec07c",       # aqua
            "Physics":"#fabd2f"   # yellow
        }
    else:
        defaults = {
            "Lecture":"#0078D7","Lab":"#E83E8C","Seminar":"#17A2B8","Exam":"#DC3545",
            "Study":"#28A745","EE":"#FD7E14","Math":"#6F42C1","CS":"#20C997","Physics":"#6610F2","Advising":"#FFC107"
        }
    for k,v in defaults.items():
        palette.setdefault(k, v)

    # Determine "now" with timezone
    if args.now:
        try:
            naive = datetime.strptime(args.now, "%Y-%m-%dT%H:%M")
        except Exception as e:
            raise SystemExit(f"--now format must be YYYY-MM-DDTHH:MM, got {args.now}")
        if args.tz and ZoneInfo:
            now_dt = naive.replace(tzinfo=ZoneInfo(args.tz))
        else:
            now_dt = naive
    else:
        if args.tz and ZoneInfo:
            now_dt = datetime.now(ZoneInfo(args.tz))
        else:
            now_dt = datetime.now()

    # Filters
    today = now_dt.date()
    if args.today:
        events = [e for e in events if e.d == today]
    elif args.this_week or ('weekly_mode' in locals() and weekly_mode):
        lo = today - timedelta(days=today.weekday())
        hi = lo + timedelta(days=6)
        events = [e for e in events if lo <= e.d <= hi]

    # Choose layout: compact if multiple days (or forced); detailed for single day
    distinct_days = sorted({e.d for e in events})
    use_compact = args.compact or (len(distinct_days) > 1)
    if use_compact:
        print_agenda_compact(events, palette, args.theme, now_dt)
    else:
        print_agenda(events, palette, args.theme, now_dt)

if __name__ == "__main__":
    main()

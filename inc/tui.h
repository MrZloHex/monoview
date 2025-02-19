#ifndef TUI_H
#define TUI_H

#include <stddef.h>

#include "kanban.h"
#include "loger.h"
#include "calendar.h"
#include "screen.h"
#include "logo.h"

typedef union
{
    Kanban   kanban;
    Loger    loger;
    Logo     logo;
    Calendar calendar;
    Screen   screen;
} View;

_Static_assert
(
    offsetof(Kanban, win)     == offsetof(Loger,    win) &&
    offsetof(Kanban, win)     == offsetof(Logo,     win) &&
    offsetof(Kanban, win)     == offsetof(Calendar, win) &&
    offsetof(Kanban, win)     == offsetof(Screen,   win),
    "WIN FIELD OFFSET MISMATCH"
);
_Static_assert
(
    offsetof(Kanban, focused) == offsetof(Loger,    focused) &&
    offsetof(Kanban, focused) == offsetof(Logo,     focused) &&
    offsetof(Kanban, focused) == offsetof(Calendar, focused) &&
    offsetof(Kanban, focused) == offsetof(Screen,   focused),
    "FOCUSED FIELD OFFSET MISMATCH"
);
_Static_assert
(
    offsetof(Kanban, width)   == offsetof(Loger,    width) &&
    offsetof(Kanban, width)   == offsetof(Logo,     width) &&
    offsetof(Kanban, width)   == offsetof(Calendar, width) &&
    offsetof(Kanban, width)   == offsetof(Screen,   width),
    "WIDTH FIELD OFFSET MISMATCH"
);
_Static_assert
(
    offsetof(Kanban, height)  == offsetof(Loger,    height) &&
    offsetof(Kanban, height)  == offsetof(Logo,     height) &&
    offsetof(Kanban, height)  == offsetof(Calendar, height) &&
    offsetof(Kanban, height)  == offsetof(Screen,   height),
    "HEIGHT FIELD OFFSET MISMATCH"
);

void init_colors();
void draw_highlight(WINDOW *win, int focused);

#endif


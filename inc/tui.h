#ifndef TUI_H
#define TUI_H

#include <stddef.h>

#include "kanban.h"
#include "loger.h"
#include "calendar.h"
#include "screen.h"
#include "logo.h"

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;
} CommonView;

typedef union
{
    CommonView view;
    Kanban     kanban;
    Loger      loger;
    Logo       logo;
    Calendar   calendar;
    Screen     screen;
} View;

_Static_assert
(
    offsetof(CommonView, win)     == offsetof(Kanban,   win) &&
    offsetof(CommonView, win)     == offsetof(Loger,    win) &&
    offsetof(CommonView, win)     == offsetof(Logo,     win) &&
    offsetof(CommonView, win)     == offsetof(Calendar, win) &&
    offsetof(CommonView, win)     == offsetof(Screen,   win),
    "WIN FIELD OFFSET MISMATCH"
);
_Static_assert
(
    offsetof(CommonView, focused) == offsetof(Kanban,   focused) &&
    offsetof(CommonView, focused) == offsetof(Loger,    focused) &&
    offsetof(CommonView, focused) == offsetof(Logo,     focused) &&
    offsetof(CommonView, focused) == offsetof(Calendar, focused) &&
    offsetof(CommonView, focused) == offsetof(Screen,   focused),
    "FOCUSED FIELD OFFSET MISMATCH"
);
_Static_assert
(
    offsetof(CommonView, width)   == offsetof(Kanban,   width) &&
    offsetof(CommonView, width)   == offsetof(Loger,    width) &&
    offsetof(CommonView, width)   == offsetof(Logo,     width) &&
    offsetof(CommonView, width)   == offsetof(Calendar, width) &&
    offsetof(CommonView, width)   == offsetof(Screen,   width),
    "WIDTH FIELD OFFSET MISMATCH"
);
_Static_assert
(
    offsetof(CommonView, height)  == offsetof(Kanban,   height) &&
    offsetof(CommonView, height)  == offsetof(Loger,    height) &&
    offsetof(CommonView, height)  == offsetof(Logo,     height) &&
    offsetof(CommonView, height)  == offsetof(Calendar, height) &&
    offsetof(CommonView, height)  == offsetof(Screen,   height),
    "HEIGHT FIELD OFFSET MISMATCH"
);

void
init_colors();

void
view_draw_focused(View view);

#endif


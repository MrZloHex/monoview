#include "tui.h"
#include <ncurses.h>

void
init_colors()
{
    if (has_colors())
    {
        start_color();
        use_default_colors();
        // Normal windows: white on black.
        init_pair(1, COLOR_WHITE, -1);
        // Focused window border: yellow on black.
        init_pair(2, COLOR_YELLOW, -1);
        // Calendar window border: cyan on black.
        init_pair(3, COLOR_CYAN, -1);
    }
}


void
view_draw_focused(View view)
{
    if (view.view.focused)
    {
        wattron(view.view.win, COLOR_PAIR(2));
        box(view.view.win, 0, 0);
        wattroff(view.view.win, COLOR_PAIR(2));
    }
    else
    {
        wattron(view.view.win, COLOR_PAIR(1));
        box(view.view.win, 0, 0);
        wattroff(view.view.win, COLOR_PAIR(1));
    }

    wrefresh(view.view.win);
}


#include "logo.h"

Logo
logo_init(int y, int x, int height, int width)
{
    // TODO: check minimal dimensions

    WINDOW *win = newwin(height, width, y, x);
    wbkgd(win, COLOR_PAIR(1));

    mvwprintw(win, 0, 4, "███╗   ███╗");
    mvwprintw(win, 1, 4, "████╗ ████║");
    mvwprintw(win, 2, 4, "██╔████╔██║");
    mvwprintw(win, 3, 4, "██║╚██╔╝██║");
    mvwprintw(win, 4, 4, "██║ ╚═╝ ██║");
    mvwprintw(win, 5, 4, "╚═╝     ╚═╝");

    wrefresh(win);


    Logo logo =
    {
        .win     = win,
        .height  = height,
        .width   = width,
        .focused = false
    };

    return logo;
}

#include "logoview.h"

WINDOW *
logoview_create(int y, int x, int height, int width)
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

    return win;
}

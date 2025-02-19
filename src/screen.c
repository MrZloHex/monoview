#include "screenview.h"
#include <time.h>

WINDOW *
screenview_create(int x, int y, int height, int width)
{
    // TODO: chekc 20x4 dims
   
    WINDOW *win = newwin(height, width, y, x);
    wbkgd(win, COLOR_PAIR(1));
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    mvwprintw(win, 1, 1, "Wednesday   19.02.25");
    mvwprintw(win, 3, 1, "-6(-11)C       R 86%%");
    mvwprintw(win, 4, 1, "      MrZloHex      ");

    screenview_update_datetime(win);

    return win;
}

void
screenview_update_datetime(WINDOW *win)
{
    time_t now = time(NULL);
    struct tm *tm_info = localtime(&now);
    char date[10];
    char time[10];
    strftime(date, sizeof(date), "%y.%m.%d", tm_info);
    strftime(time, sizeof(time), "%H:%M:%S", tm_info);
    mvwprintw(win, 1, 13, "%s", date);
    mvwprintw(win, 2, 1,  "C     %s     S", time);
    wrefresh(win);
}

#include "screen.h"

#include <time.h>

Screen
screen_init(int x, int y, int height, int width)
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

    wrefresh(win);

    Screen scr =
    {
        .win     = win,
        .focused = false,
        .height  = height,
        .width   = width
    };

    return scr;
}

void
screen_update_datetime(Screen *scr)
{
    time_t now = time(NULL);
    struct tm *tm_info = localtime(&now);
    char date[10];
    char time[10];
    strftime(date, sizeof(date), "%y.%m.%d", tm_info);
    strftime(time, sizeof(time), "%H:%M:%S", tm_info);
    mvwprintw(scr->win, 1, 13, "%s", date);
    mvwprintw(scr->win, 2, 1,  "C     %s     S", time);
    wrefresh(scr->win);
}

#include "calendar.h"
#include <string.h>


Calendar
calendar_init(int starty, int startx, int height, int width)
{
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(3));
    wattron(win, COLOR_PAIR(3));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(3));
    mvwprintw(win, 1, (width - 17) / 2, "University Schedule");

    int available_width = width - 2;
    int day_width = available_width / 7;
    const char *days[7] = {"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"};
    for (int i = 0; i < 7; i++) {
        int col = 1 + i * day_width;
        int offset = (day_width - strlen(days[i])) / 2;
        mvwprintw(win, 2, col + offset, "%s", days[i]);
    }
    int col = 1 + 1 * day_width;
    mvwprintw(win, 7, col, "15:30 CH");
    mvwprintw(win, 9, col, "18:35 DB");
    col = 1 + 2 * day_width;
    mvwprintw(win, 6, col, "13:55 DE");
    mvwprintw(win, 7, col, "15:30 PT");
    col = 1 + 3 * day_width;
    mvwprintw(win, 4, col, "10:45 HA");
    mvwprintw(win, 5, col, "12:20 HA");
    mvwprintw(win, 8, col, "17:05 DE");
    col = 1 + 4 * day_width;
    mvwprintw(win, 5, col, "12:20 CC");
    mvwprintw(win, 8, col, "17:05 DM");
    mvwprintw(win, 9, col, "18:30 CH");
    col = 1 + 5 * day_width;
    mvwprintw(win, 4, col, "10:45 AD");
    mvwprintw(win, 5, col, "12:20 AD");

    wrefresh(win);

    Calendar cal =
    {
        .win     = win,
        .height  = height,
        .width   = width,
        .focused = false
    };


    return cal;
}

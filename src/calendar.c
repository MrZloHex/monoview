#include "calendar.h"
#include <string.h>


WINDOW *create_calendar_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(3));
    wattron(win, COLOR_PAIR(3));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(3));
    mvwprintw(win, 1, (width - 17) / 2, "University Schedule");

    int available_width = width - 2; // inside border
    int day_width = available_width / 7;
    const char *days[7] = {"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"};
    for (int i = 0; i < 7; i++) {
        int col = 1 + i * day_width;
        int offset = (day_width - strlen(days[i])) / 2;
        mvwprintw(win, 2, col + offset, "%s", days[i]);
    }
    // Sample schedule entries.
    for (int i = 0; i < 7; i++) {
        int col = 1 + i * day_width;
        mvwprintw(win, 3, col, "09:00 Lec");
        mvwprintw(win, 4, col, "11:00 Lab");
        mvwprintw(win, 5, col, "13:00 Sem");
        mvwprintw(win, 6, col, "HW: Calc");
        mvwprintw(win, 7, col, "HW: Phys");
    }
    return win;
}

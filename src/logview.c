#include "logview.h"

WINDOW *create_logs_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    mvwprintw(win, 1, 1, "Logs:");
    mvwprintw(win, 2, 1, "Msg: Connected");
    mvwprintw(win, 3, 1, "Msg: Data received");
    mvwprintw(win, 4, 1, "Msg: No errors");
    return win;
}

#include "canban.h"

WINDOW *create_diary_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    return win;
}

void update_diary_window(WINDOW *win, int height, int width) {
    werase(win);
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    int col_width = (width - 2) / 3;
    // Column headers.
    mvwprintw(win, 1, 1 + (col_width - 6) / 2, "To Do");
    mvwprintw(win, 1, 1 + col_width + (col_width - 10) / 2, "In Prog");
    mvwprintw(win, 1, 1 + 2 * col_width + (col_width - 4) / 2, "Done");
    for (int y = 1; y < height - 1; y++) {
        mvwaddch(win, y, col_width, ACS_VLINE);
        mvwaddch(win, y, col_width * 2, ACS_VLINE);
    }
    int line = 3;
    for (int i = 0; i < todo_count && line < height - 1; i++, line++) {
        mvwprintw(win, line, 1, "- %s", todo_list[i]);
    }
    wrefresh(win);
}

#include "ncurses.h"

WINDOW *huy(int starty, int startx) {
    int height = 4 + 2;  // 4 rows for "LCD" content plus 2 for border
    int width = 20 + 2;  // 20 columns plus 2 for border
    WINDOW *win = newwin(height, width, starty, startx);
    box(win, 0, 0);
    mvwprintw(win, 1, 1, "Wednesday   19.02.25");
    mvwprintw(win, 2, 1, "C     01:54:08     S");
    mvwprintw(win, 3, 1, "-6(-11)C       R 86%%");
    mvwprintw(win, 4, 1, "      MrZloHex      ");
    return win;
}

#include "tui.h"
#include <ncurses.h>

void init_colors() {
    if (has_colors()) {
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


void draw_highlight(WINDOW *win, int focused) {
    if (focused) {
        wattron(win, COLOR_PAIR(2));
        box(win, 0, 0);
        wattroff(win, COLOR_PAIR(2));
    } else {
        wattron(win, COLOR_PAIR(1));
        box(win, 0, 0);
        wattroff(win, COLOR_PAIR(1));
    }
    wrefresh(win);
}


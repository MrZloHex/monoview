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

#include <wchar.h>

WINDOW *create_logo_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(1)); // use default color pair (or your chosen one)
    wattron(win, COLOR_PAIR(1));
    // Print the ASCII art logo, starting at row 1, col 1 (inside the border)
    mvwprintw(win, 0, 4, " ███╗   ███╗");
    mvwprintw(win, 1, 4, " ████╗ ████║");
    mvwprintw(win, 2, 4, " ██╔████╔██║");
    mvwprintw(win, 3, 4, " ██║╚██╔╝██║ ");
    mvwprintw(win, 4, 4, " ██║ ╚═╝ ██║");
    mvwprintw(win, 5, 4, " ╚═╝     ╚═╝");

    wrefresh(win);
    return win;
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


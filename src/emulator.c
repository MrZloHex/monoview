#include "emulator.h"
#include <time.h>

WINDOW *create_lcd_window(int starty, int startx) {
    WINDOW *win = newwin(LCD_HEIGHT, LCD_WIDTH, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    mvwprintw(win, 1, 1, "Wednesday   19.02.25");
    mvwprintw(win, 3, 1, "-6(-11)C       R 86%%");
    mvwprintw(win, 4, 1, "      MrZloHex      ");
    return win;
}

void update_lcd_time(WINDOW *lcd_win) {
    time_t now = time(NULL);
    struct tm *tm_info = localtime(&now);
    char time_str[9];
    strftime(time_str, sizeof(time_str), "%H:%M:%S", tm_info);
    mvwprintw(lcd_win, 2, 1, "C     %s     S", time_str);
    wrefresh(lcd_win);
}

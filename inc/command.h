#ifndef COMMAND_H
#define COMMAND_H

#include <ncurses.h>
void enter_command_mode(int max_y, int max_x, WINDOW *vert_sep_top, WINDOW *vert_sep_bottom,
    WINDOW *horz_sep, WINDOW **win_array, int num_windows,
    WINDOW *diary_win, int diary_height, int diary_width,
    WINDOW *lcd_win, WINDOW *logo_win);

void enter_new_entry(WINDOW *diary_win, int diary_height, int diary_width);

#endif


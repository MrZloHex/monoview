/* ==============================================================================
 *
 *  ███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
 *  ████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
 *  ██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
 *  ██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
 *  ██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
 *  ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝
 *
 *                           ░▒▓█ _MonoView_ █▓▒░
 *
 *   File       : main.c
 *   Author     : MrZloHex
 *   Date       : 2025-02-19
 *
 * ==============================================================================
 */






#include <ncurses.h>
#include <stdlib.h>
#include "tui.h"
#include "command.h"

char todo_list[MAX_TODO][256];
int todo_count = 0;
int command_quit = 0;

int main() {
    initscr();
    cbreak();
    noecho();
    keypad(stdscr, TRUE);
    curs_set(0);

    init_colors();
    bkgd(COLOR_PAIR(1));
    refresh();

    timeout(500);
    int max_y, max_x;
    getmaxyx(stdscr, max_y, max_x);

    if (max_y < (LCD_HEIGHT + LOGO_HEIGHT + 5) || max_x < (LCD_WIDTH + LOGS_WIDTH + 20)) {
        endwin();
        printf("Terminal too small. Please resize.\n");
        return 1;
    }

    int left_top_total = LCD_HEIGHT + LOGO_HEIGHT;
    int top_row_height = (left_top_total > CALENDAR_HEIGHT ? left_top_total : CALENDAR_HEIGHT);
    int bottom_row_height = max_y - top_row_height;

    // Create separators.
    WINDOW *horz_sep = newwin(1, max_x, top_row_height, 0);
    for (int x = 0; x < max_x; x++) {
        mvwaddch(horz_sep, 0, x, ACS_HLINE);
    }
    wrefresh(horz_sep);

    WINDOW *vert_sep_top = newwin(top_row_height, 1, 0, LCD_WIDTH);
    for (int y = 0; y < top_row_height; y++) {
        mvwaddch(vert_sep_top, y, 0, ACS_VLINE);
    }
    wrefresh(vert_sep_top);

    WINDOW *vert_sep_bottom = newwin(bottom_row_height, 1, top_row_height, LOGS_WIDTH);
    for (int y = 0; y < bottom_row_height; y++) {
        mvwaddch(vert_sep_bottom, y, 0, ACS_VLINE);
    }
    wrefresh(vert_sep_bottom);

    // Create windows.
    WINDOW *lcd_win = create_lcd_window(0, 0);
    WINDOW *logo_win = create_logo_window(LCD_HEIGHT, 0, LOGO_HEIGHT, LCD_WIDTH);
    WINDOW *calendar_win = create_calendar_window(0, LCD_WIDTH + 1, CALENDAR_HEIGHT, max_x - LCD_WIDTH - 1);

    WINDOW *logs_win = create_logs_window(top_row_height, 0, bottom_row_height, LOGS_WIDTH);
    WINDOW *diary_win = create_diary_window(top_row_height, LOGS_WIDTH + 1, bottom_row_height, max_x - LOGS_WIDTH - 1);
    update_diary_window(diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1);

    WINDOW *win_array[NUM_WINDOWS] = { lcd_win, calendar_win, logs_win, diary_win };
    int current_focus = 0;

    for (int i = 0; i < NUM_WINDOWS; i++) {
        draw_highlight(win_array[i], (i == current_focus));
    }

    int ch;
    while ((ch = getch()) != 'q' && !command_quit) {
        if (ch == ERR) {
            update_lcd_time(lcd_win);
            continue;
        }
        if (ch == '\t' || ch == KEY_BTAB) {
            current_focus = (current_focus + 1) % NUM_WINDOWS;
            for (int i = 0; i < NUM_WINDOWS; i++) {
                draw_highlight(win_array[i], (i == current_focus));
            }
        } else if (ch == 'h') {
            if (current_focus == 1) current_focus = 0;
            else if (current_focus == 3) current_focus = 2;
            for (int i = 0; i < NUM_WINDOWS; i++) {
                draw_highlight(win_array[i], (i == current_focus));
            }
        } else if (ch == 'l') {
            if (current_focus == 0) current_focus = 1;
            else if (current_focus == 2) current_focus = 3;
            for (int i = 0; i < NUM_WINDOWS; i++) {
                draw_highlight(win_array[i], (i == current_focus));
            }
        } else if (ch == 'j') {
            if (current_focus == 0) current_focus = 2;
            else if (current_focus == 1) current_focus = 3;
            for (int i = 0; i < NUM_WINDOWS; i++) {
                draw_highlight(win_array[i], (i == current_focus));
            }
        } else if (ch == 'k') {
            if (current_focus == 2) current_focus = 0;
            else if (current_focus == 3) current_focus = 1;
            for (int i = 0; i < NUM_WINDOWS; i++) {
                draw_highlight(win_array[i], (i == current_focus));
            }
        } else if (ch == ':') {
            enter_command_mode(max_y, max_x, vert_sep_top, vert_sep_bottom, horz_sep,
                               win_array, NUM_WINDOWS, diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1,
                               lcd_win, logo_win);
        }
    }

    delwin(lcd_win);
    delwin(logo_win);
    delwin(calendar_win);
    delwin(logs_win);
    delwin(diary_win);
    delwin(vert_sep_top);
    delwin(vert_sep_bottom);
    delwin(horz_sep);
    endwin();
    return 0;
}



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
#include <locale.h>

#include <ncurses.h>
#include <stdlib.h>
#include "setup.h"
#include "command.h"


int command_quit = 0;



#include "weekview.h"
#include "logoview.h"
#include "screenview.h"

int main() {

    setlocale(LC_ALL, "");
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

    // Create windows.
    WINDOW *logo = logoview_create(0, 0, LOGO_HEIGHT, LCD_WIDTH);
    WINDOW *screen = screenview_create(0, LCD_HEIGHT, LCD_HEIGHT, LCD_WIDTH);
    WINDOW *calendar_win = create_calendar_window(0, LCD_WIDTH + 1, CALENDAR_HEIGHT, max_x - LCD_WIDTH - 1);
    LogView *log = logview_init(top_row_height, 0, bottom_row_height, LOGS_WIDTH);
    WINDOW *logs_win = log->win;
    Kanban *kan = kanban_init(top_row_height, LOGS_WIDTH + 1, bottom_row_height, max_x - LOGS_WIDTH - 1);
    WINDOW *diary_win = kan->win;
    kanban_update(diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1);

    // Focusable windows: LCD, Calendar, Logs, Diary.
    WINDOW *win_array[NUM_WINDOWS] = { screen, calendar_win, logs_win, diary_win };
    int current_focus = 0;
    for (int i = 0; i < NUM_WINDOWS; i++) {
        draw_highlight(win_array[i], (i == current_focus));
    }

    char buffer[32];
    int ch;
    while ((ch = getch()) != 'q' && !command_quit) {
        if (ch == ERR) {
            screenview_update_datetime(screen);
            update_logs_window(logs_win, bottom_row_height, LOGS_WIDTH);
            continue;
        }
        if (ch == '\t' || ch == KEY_BTAB) {
            snprintf(buffer, sizeof(buffer), "NEXT WINDOW %d", current_focus);
            log_action(buffer);
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
            //enter_command_mode(max_y, max_x, vert_sep_top, vert_sep_bottom, horz_sep,
            //                   win_array, NUM_WINDOWS, diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1,
            //                   lcd_win, logo_win);
        } else if (ch == 'i') {
            // When focused on the Diary (Kanban) window, press 'i' to add a new entry.
            if (current_focus == 3) {
                enter_new_entry(diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1);
            }
        }

        if (current_focus == 2) {  // Assuming logs window is index 2 in win_array
            if (ch == KEY_UP) {
                if (log_scroll_offset > 0) {
                    log_scroll_offset--;
                    update_logs_window(logs_win, bottom_row_height, LOGS_WIDTH);
                }
                continue;  // Prevent grid navigation in this case.
            } else if (ch == KEY_DOWN) {
                // Only scroll if there are more lines to show.
                if (log_scroll_offset < num_log_lines - (bottom_row_height - 2)) {
                    log_scroll_offset++;
                    update_logs_window(logs_win, bottom_row_height, LOGS_WIDTH);
                }
                continue;
            }
        }

    }

    delwin(screen);
    delwin(logo);
    delwin(calendar_win);
    delwin(logs_win);
    delwin(diary_win);
    endwin();
    return 0;
}


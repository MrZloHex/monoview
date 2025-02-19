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


int main() {

    setlocale(LC_ALL, "");
    initscr();
    raw();
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


    View logo     = (View)logo_init(0, 0, LOGO_HEIGHT, LCD_WIDTH);
    View screen   = (View)screen_init(0, LCD_HEIGHT, LCD_HEIGHT, LCD_WIDTH);
    View loger    = (View)loger_init(top_row_height, 0, bottom_row_height, LOGS_WIDTH);
    View calendar = (View)calendar_init(0, LCD_WIDTH +1, CALENDAR_HEIGHT, max_x - LCD_WIDTH -1);
    View kanban   = (View)kanban_init(top_row_height, LOGS_WIDTH +1, bottom_row_height, max_x - LOGS_WIDTH -1);

    #define VIEW_FOCUS_SIZE 3
    View views[VIEW_FOCUS_SIZE] = { calendar, kanban, loger};
    size_t focus = 0;
    views[focus].view.focused = true;
    for (size_t i = 0; i < VIEW_FOCUS_SIZE; ++i)
    {
        view_draw_focused(views[i]);
    }

    char buffer[32];

    log_action("HUY");
    bool run = true;
    while (run)
    {
        int ch = getch();
        if (ch == ERR) {
            screen_update_datetime(&screen.screen);
            log_action("HUY");
            loger_update(&loger.loger);
            continue;
        }
        if (ch == '\t' || ch == KEY_BTAB) {
            focus = (focus + 1) % VIEW_FOCUS_SIZE;
            snprintf(buffer, sizeof(buffer), "NEXT WINDOW %zu", focus);
            log_action(buffer);
            for (size_t i = 0; i < VIEW_FOCUS_SIZE; ++i)
            {
                views[i].view.focused = (i == focus);
                view_draw_focused(views[i]);
            }
        }
        else if (ch == 'q') 
        {
            run = false;
        } else if (ch == ':') {
            //enter_command_mode(max_y, max_x, vert_sep_top, vert_sep_bottom, horz_sep,
            //                   win_array, NUM_WINDOWS, diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1,
            //                   lcd_win, logo_win);
        } else if (ch == 'i') {
            // When focused on the Diary (Kanban) window, press 'i' to add a new entry.
            if (focus == 3) {
                //enter_new_entry(diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1);
            }
        }

        if (focus == 2)
        {
            if (ch == KEY_UP) {
                if (log_scroll_offset > 0) {
                    log_scroll_offset--;
                    loger_update(&loger.loger);
                }
                continue;  // Prevent grid navigation in this case.
            } else if (ch == KEY_DOWN) {
                // Only scroll if there are more lines to show.
                if (log_scroll_offset < num_log_lines - (bottom_row_height - 2)) {
                    log_scroll_offset++;
                    loger_update(&loger.loger);
                }
                continue;
            }
        }

    }

    delwin(screen.view.win);
    delwin(logo.view.win);
    delwin(calendar.view.win);
    delwin(loger.view.win);
    delwin(kanban.view.win);
    endwin();
    return 0;
}


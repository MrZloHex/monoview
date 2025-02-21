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

#include <locale.h>

#include "tui.h"
#include "setup.h"
#include "client.h"


int
main()
{
    if (start_ws_thread("127.0.0.1", 8080, "/") != 0) {
        fprintf(stderr, "WebSocket thread failed to start.\n");
        return EXIT_FAILURE;
    }

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

    if (max_y < (LCD_HEIGHT + LOGO_HEIGHT + 5) || max_x < (LCD_WIDTH + LOGS_WIDTH + 20))
    {
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
    kanban_update(&kanban.kanban);

    #define VIEW_FOCUS_SIZE 3
    View *views[VIEW_FOCUS_SIZE] = { &calendar, &kanban, &loger };
    size_t focus = 0;
    views[focus]->view.focused = true;
    for (size_t i = 0; i < VIEW_FOCUS_SIZE; ++i)
    {
        view_draw_focused(*views[i]);
    }

    char buffer[32];

    bool run = true;
    while (run)
    {
        int ch = getch();
        if (ch == ERR)
        {
            screen_update_datetime(&screen.screen);
            continue;
        }

        if (ch == '\t' || ch == KEY_BTAB)
        {
            focus = (focus + 1) % VIEW_FOCUS_SIZE;
            snprintf(buffer, sizeof(buffer), "NEXT WINDOW %zu", focus);
            log_action(&loger.loger, buffer);
            for (size_t i = 0; i < VIEW_FOCUS_SIZE; ++i)
            {
                views[i]->view.focused = (i == focus);
                snprintf(buffer, sizeof(buffer), "%zu: %u", i, views[i]->view.focused);
                log_action(&loger.loger, buffer);
                view_draw_focused(*views[i]);
            }
        }
        else if (ch == 'q') 
        {
            run = false;
        }

        if (kanban.kanban.focused)
        {
            kanban_pressed(&kanban.kanban, ch);
        }

        if (loger.loger.focused)
        {
            loger_pressed(&loger.loger, ch);
        }

    }

    delwin(screen.view.win);
    delwin(logo.view.win);
    delwin(calendar.view.win);
    delwin(loger.view.win);
    delwin(kanban.view.win);
    endwin();

    stop_ws_thread();
    return 0;
}


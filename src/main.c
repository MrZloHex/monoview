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

// Fixed dimensions for the left column (LCD and Logs)
#define LEFT_WIDTH 22
// Fixed height for the top area (LCD and Calendar)
#define TOP_HEIGHT 6
// Fixed height for the CMD LINE window at the bottom right
#define CMDLINE_HEIGHT 3

// Modular function to create the LCD window (emulated 20x4 plus borders)
WINDOW *create_lcd_window(int starty, int startx) {
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

// Modular function to create the Calendar window
WINDOW *create_calendar_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    box(win, 0, 0);
    mvwprintw(win, 1, 1, "Calendar:");
    mvwprintw(win, 2, 1, "2025-02-19");
    mvwprintw(win, 3, 1, "Meeting at 10:00");
    return win;
}

// Modular function to create the Logs window
WINDOW *create_logs_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    box(win, 0, 0);
    mvwprintw(win, 1, 1, "Logs:");
    mvwprintw(win, 2, 1, "Connected");
    mvwprintw(win, 3, 1, "Data received");
    mvwprintw(win, 4, 1, "No errors");
    // You can add more log entries as needed
    return win;
}

// Modular function to create the Diary window
WINDOW *create_diary_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    box(win, 0, 0);
    mvwprintw(win, 1, 1, "Diary:");
    mvwprintw(win, 2, 1, "Started project");
    mvwprintw(win, 3, 1, "Initial tests OK");
    mvwprintw(win, 4, 1, "Waiting for input...");
    return win;
}

// Modular function to create the CMD LINE window
WINDOW *create_cmdline_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    box(win, 0, 0);
    mvwprintw(win, 1, 1, "CMD: ");
    return win;
}

int main() {
    // Initialize ncurses
    initscr();
    cbreak();
    noecho();
    curs_set(0);
    refresh();

    int max_y, max_x;
    getmaxyx(stdscr, max_y, max_x);

    // Calculate dimensions for the right column
    int right_width = max_x - LEFT_WIDTH;
    int bottom_height = max_y - TOP_HEIGHT;

    // Top area: LCD on the left and Calendar on the right.
    WINDOW *lcd_win = create_lcd_window(0, 0);
    WINDOW *calendar_win = create_calendar_window(0, LEFT_WIDTH, TOP_HEIGHT, right_width);

    // Bottom left: Logs window (full height of bottom area, fixed width)
    WINDOW *logs_win = create_logs_window(TOP_HEIGHT, 0, bottom_height, LEFT_WIDTH);

    // Bottom right: Split into Diary (upper part) and CMD Line (lower part)
    int diary_height = bottom_height - CMDLINE_HEIGHT;
    WINDOW *diary_win = create_diary_window(TOP_HEIGHT, LEFT_WIDTH, diary_height, right_width);
    WINDOW *cmdline_win = create_cmdline_window(TOP_HEIGHT + diary_height, LEFT_WIDTH, CMDLINE_HEIGHT, right_width);

    // Refresh all windows
    wrefresh(lcd_win);
    wrefresh(calendar_win);
    wrefresh(logs_win);
    wrefresh(diary_win);
    wrefresh(cmdline_win);

    // Optionally print an instruction at the bottom of the screen
    mvprintw(max_y - 1, 1, "Press any key to exit...");
    refresh();

    // Wait for user input
    getch();

    // Cleanup all windows and exit ncurses mode
    delwin(lcd_win);
    delwin(calendar_win);
    delwin(logs_win);
    delwin(diary_win);
    delwin(cmdline_win);
    endwin();

    return 0;
}


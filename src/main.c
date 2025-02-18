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
#include <string.h>
#include <time.h>

#define LCD_HEIGHT 6         // LCD: 4 rows content + 2 for border
#define LCD_WIDTH 22         // LCD window width (emulated 20x4)
#define LOGO_HEIGHT 3        // Logo window height under LCD
#define CALENDAR_HEIGHT 12   // Calendar window height (for schedule details)
#define LOGS_WIDTH 40        // Logs window width (wider)
#define NUM_WINDOWS 4        // Focusable windows: LCD, Calendar, Logs, Diary
#define MAX_TODO 100         // Maximum number of tasks

// Global ToDo list (for the Kanban "To Do" column)
char todo_list[MAX_TODO][256];
int todo_count = 0;
int command_quit = 0;        // Global flag: set if "quit" command is issued

// --- Color Initialization --- //
void init_colors() {
    if (has_colors()) {
        start_color();
        // Normal windows: white on black.
        init_pair(1, COLOR_WHITE, COLOR_BLACK);
        // Focused window border: yellow on black.
        init_pair(2, COLOR_YELLOW, COLOR_BLACK);
        // Calendar window border: cyan on black.
        init_pair(3, COLOR_CYAN, COLOR_BLACK);
    }
}

// --- Modular Window Creation Functions --- //

// LCD window: shows header and current time.
WINDOW *create_lcd_window(int starty, int startx) {
    WINDOW *win = newwin(LCD_HEIGHT, LCD_WIDTH, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    mvwprintw(win, 1, 1, "LCD Emulation:");
    return win;
}

// Logo window: placed immediately below the LCD window.
WINDOW *create_logo_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    // Center a simple logo message.
    mvwprintw(win, height/2, (width - 9)/2, "MyProject");
    return win;
}

// Calendar window: displays a horizontal list of days with sample schedule entries.
WINDOW *create_calendar_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(3));
    wattron(win, COLOR_PAIR(3));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(3));
    mvwprintw(win, 1, (width - 17) / 2, "University Schedule");

    int available_width = width - 2; // inside border
    int day_width = available_width / 7;
    const char *days[7] = {"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"};

    // Print day names on row 2.
    for (int i = 0; i < 7; i++) {
        int col = 1 + i * day_width;
        int offset = (day_width - (int)strlen(days[i])) / 2;
        mvwprintw(win, 2, col + offset, "%s", days[i]);
    }
    // Sample schedule entries.
    for (int i = 0; i < 7; i++) {
        int col = 1 + i * day_width;
        mvwprintw(win, 3, col, "09:00 Lec");
        mvwprintw(win, 4, col, "11:00 Lab");
        mvwprintw(win, 5, col, "13:00 Sem");
        mvwprintw(win, 6, col, "HW: Calc");
        mvwprintw(win, 7, col, "HW: Phys");
    }
    return win;
}

// Logs window: displays communication logs.
WINDOW *create_logs_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    mvwprintw(win, 1, 1, "Logs:");
    mvwprintw(win, 2, 1, "Msg: Connected");
    mvwprintw(win, 3, 1, "Msg: Data received");
    mvwprintw(win, 4, 1, "Msg: No errors");
    return win;
}

// Diary (Kanban) window: displays a simple board with a "To Do" column.
WINDOW *create_diary_window(int starty, int startx, int height, int width) {
    WINDOW *win = newwin(height, width, starty, startx);
    wbkgd(win, COLOR_PAIR(1));
    return win;
}

// Update the Diary window with the current "To Do" list.
void update_diary_window(WINDOW *win, int height, int width) {
    werase(win);
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    int col_width = (width - 2) / 3;
    // Column headers.
    mvwprintw(win, 1, 1 + (col_width - 6) / 2, "To Do");
    mvwprintw(win, 1, 1 + col_width + (col_width - 10) / 2, "In Prog");
    mvwprintw(win, 1, 1 + 2 * col_width + (col_width - 4) / 2, "Done");

    // Draw vertical separators.
    for (int y = 1; y < height - 1; y++) {
        mvwaddch(win, y, col_width, ACS_VLINE);
        mvwaddch(win, y, col_width * 2, ACS_VLINE);
    }
    // Print tasks in the "To Do" column.
    int line = 3;
    for (int i = 0; i < todo_count && line < height - 1; i++, line++) {
        mvwprintw(win, line, 1, "- %s", todo_list[i]);
    }
    wrefresh(win);
}

// --- Helper Function to Draw a Focused Border Using Colors --- //
// Focused windows are drawn with color pair 2.
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

// Update the LCD window to display the current time.
void update_lcd_time(WINDOW *lcd_win) {
    time_t now = time(NULL);
    struct tm *tm_info = localtime(&now);
    char time_str[9]; // Format: HH:MM:SS
    strftime(time_str, sizeof(time_str), "%H:%M:%S", tm_info);
    mvwprintw(lcd_win, 3, 1, "Time: %s", time_str);
    wclrtoeol(lcd_win);
    wrefresh(lcd_win);
}

// --- Non-blocking Command Mode --- //
// A temporary command window appears at the bottom. It uses its own loop (with nodelay)
// so that the LCD clock (and any other updates) continue. This loop collects input
// until Enter is pressed, then processes "add <task>" and "quit" commands.
// After finishing, it removes itself and forces a full redraw of the layout.
void enter_command_mode(int max_y, int max_x, WINDOW *vert_sep_top, WINDOW *vert_sep_bottom,
                          WINDOW *horz_sep, WINDOW **win_array, int num_windows,
                          WINDOW *diary_win, int diary_height, int diary_width,
                          WINDOW *lcd_win, WINDOW *logo_win) {
    int cmd_height = 3;
    WINDOW *cmd_win = newwin(cmd_height, max_x, max_y - cmd_height, 0);
    wbkgd(cmd_win, COLOR_PAIR(1));
    box(cmd_win, 0, 0);
    mvwprintw(cmd_win, 1, 1, "CMD: ");
    wrefresh(cmd_win);
    // Set non-blocking mode.
    nodelay(cmd_win, TRUE);

    char cmd[80];
    int idx = 0;
    memset(cmd, 0, sizeof(cmd));
    int ch;

    // Loop until Enter is pressed.
    while (1) {
        update_lcd_time(lcd_win);
        ch = wgetch(cmd_win);
        if (ch != ERR) {
            if (ch == '\n' || ch == KEY_ENTER) {
                break;
            } else if (ch == KEY_BACKSPACE || ch == 127 || ch == '\b') {
                if (idx > 0) {
                    idx--;
                    cmd[idx] = '\0';
                }
            } else {
                if (idx < 79) {
                    cmd[idx++] = ch;
                    cmd[idx] = '\0';
                }
            }
            werase(cmd_win);
            box(cmd_win, 0, 0);
            mvwprintw(cmd_win, 1, 1, "CMD: %s", cmd);
            wrefresh(cmd_win);
        }
        napms(100);
    }

    // Process command.
    if (strncmp(cmd, "add ", 4) == 0) {
        char *task = cmd + 4;
        if (todo_count < MAX_TODO) {
            strncpy(todo_list[todo_count], task, 255);
            todo_list[todo_count][255] = '\0';
            todo_count++;
        }
    }
    if (strncmp(cmd, "quit", 4) == 0) {
        command_quit = 1;
    }

    // Clear and remove the command window.
    werase(cmd_win);
    wrefresh(cmd_win);
    delwin(cmd_win);

    // Redraw the static layout to remove any leftover white space.
    clear();
    refresh();
    wrefresh(vert_sep_top);
    wrefresh(vert_sep_bottom);
    wrefresh(horz_sep);
    for (int i = 0; i < num_windows; i++) {
        touchwin(win_array[i]);
        wrefresh(win_array[i]);
    }
    touchwin(logo_win);
    wrefresh(logo_win);
    update_diary_window(diary_win, diary_height, diary_width);
}

int main() {
    initscr();
    cbreak();
    noecho();
    keypad(stdscr, TRUE);
    curs_set(0);

    init_colors();
    bkgd(COLOR_PAIR(1)); // Set default background.
    refresh();

    // Set timeout for periodic updates.
    timeout(500); // 500 ms

    int max_y, max_x;
    getmaxyx(stdscr, max_y, max_x);

    if (max_y < (LCD_HEIGHT + LOGO_HEIGHT + 5) || max_x < (LCD_WIDTH + LOGS_WIDTH + 20)) {
        endwin();
        printf("Terminal too small. Please resize.\n");
        return 1;
    }

    // Calculate top row height.
    int left_top_total = LCD_HEIGHT + LOGO_HEIGHT;
    int top_row_height = (left_top_total > CALENDAR_HEIGHT ? left_top_total : CALENDAR_HEIGHT);
    int bottom_row_height = max_y - top_row_height;

    // --- Create Separators --- //
    // Horizontal separator between top and bottom rows.
    WINDOW *horz_sep = newwin(1, max_x, top_row_height, 0);
    for (int x = 0; x < max_x; x++) {
        mvwaddch(horz_sep, 0, x, ACS_HLINE);
    }
    wrefresh(horz_sep);

    // Vertical separator in top row between LCD/logo and Calendar at col = LCD_WIDTH.
    WINDOW *vert_sep_top = newwin(top_row_height, 1, 0, LCD_WIDTH);
    for (int y = 0; y < top_row_height; y++) {
        mvwaddch(vert_sep_top, y, 0, ACS_VLINE);
    }
    wrefresh(vert_sep_top);

    // Vertical separator in bottom row between Logs and Diary at col = LOGS_WIDTH.
    WINDOW *vert_sep_bottom = newwin(bottom_row_height, 1, top_row_height, LOGS_WIDTH);
    for (int y = 0; y < bottom_row_height; y++) {
        mvwaddch(vert_sep_bottom, y, 0, ACS_VLINE);
    }
    wrefresh(vert_sep_bottom);

    // --- Create Windows --- //
    // Top row: left column contains LCD and Logo.
    WINDOW *lcd_win = create_lcd_window(0, 0);
    WINDOW *logo_win = create_logo_window(LCD_HEIGHT, 0, LOGO_HEIGHT, LCD_WIDTH);
    WINDOW *calendar_win = create_calendar_window(0, LCD_WIDTH + 1, CALENDAR_HEIGHT, max_x - LCD_WIDTH - 1);

    // Bottom row:
    WINDOW *logs_win = create_logs_window(top_row_height, 0, bottom_row_height, LOGS_WIDTH);
    WINDOW *diary_win = create_diary_window(top_row_height, LOGS_WIDTH + 1, bottom_row_height, max_x - LOGS_WIDTH - 1);
    update_diary_window(diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1);

    // Focusable windows array (order: LCD, Calendar, Logs, Diary).
    WINDOW *win_array[NUM_WINDOWS] = { lcd_win, calendar_win, logs_win, diary_win };
    int current_focus = 0;

    // Initial focus highlight.
    for (int i = 0; i < NUM_WINDOWS; i++) {
        draw_highlight(win_array[i], (i == current_focus));
    }

    int ch;
    while ((ch = getch()) != 'q' && !command_quit) {
        if (ch == ERR) {
            update_lcd_time(lcd_win);
            continue;
        }
        // Cycle focus with Tab.
        if (ch == '\t' || ch == KEY_BTAB) {
            current_focus = (current_focus + 1) % NUM_WINDOWS;
            for (int i = 0; i < NUM_WINDOWS; i++) {
                draw_highlight(win_array[i], (i == current_focus));
            }
        }
        // Vim-like motion keys.
        else if (ch == 'h') {
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
        }
        else if (ch == ':') {
            enter_command_mode(max_y, max_x, vert_sep_top, vert_sep_bottom, horz_sep,
                               win_array, NUM_WINDOWS, diary_win, bottom_row_height, max_x - LOGS_WIDTH - 1,
                               lcd_win, logo_win);
        }
    }

    // --- Cleanup ---
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


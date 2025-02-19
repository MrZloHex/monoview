#include "loger.h"
#include <string.h>
#include <stdlib.h>


char log_lines[MAX_LOG_LINES][MAX_LOG_LENGTH] = {0};
int num_log_lines = 0;
int log_scroll_offset = 0;


void log_action(const char *action) {
    if (num_log_lines < MAX_LOG_LINES) {
        strncpy(log_lines[num_log_lines], action, MAX_LOG_LENGTH - 1);
        log_lines[num_log_lines][MAX_LOG_LENGTH - 1] = '\0';
        num_log_lines++;
    }
}

Loger
loger_init(int y, int x, int height, int width)
{
    Loger log;
    // TODO: check it
    
    log.focused   = 0;
    log.q_entries = 0;

    log.win = newwin(height, width, y, x);
    wbkgd(log.win, COLOR_PAIR(1));
    wattron(log.win, COLOR_PAIR(1));
    box(log.win, 0, 0);
    wattroff(log.win, COLOR_PAIR(1));
    return log;
}


void update_logs_window(WINDOW *win, int height, int width) {
    werase(win);
    wbkgd(win, COLOR_PAIR(1));
    box(win, 0, 0);
    // The inner height available (excluding top and bottom borders)
    int available = height - 2;
    int start_line = log_scroll_offset;
    int end_line = (num_log_lines < start_line + available) ? num_log_lines : start_line + available;
    int line = 1;
    // Print log lines within the inner area (columns 1 .. width-3, leaving column width-2 for scroll bar)
    for (int i = start_line; i < end_line; i++, line++) {
        // Print only within columns 1 to (width - 3)
        mvwprintw(win, line, 1, "[%u] %.*s", i, width - 3, log_lines[i]);
    }

    // Draw a scroll bar if necessary.
    if (num_log_lines > available) {
        // Calculate the scrollbar height proportionally.
        int scrollbar_height = (available * available) / num_log_lines;
        if (scrollbar_height < 1)
            scrollbar_height = 1;
        int max_offset = num_log_lines - available;
        int scrollbar_start = 0;
        if (max_offset > 0)
            scrollbar_start = (log_scroll_offset * (available - scrollbar_height)) / max_offset;
        // Choose the column for the scroll bar; here we use the second-to-last column (inside the right border).
        int scroll_col = width - 2;
        // Clear that column (within inner area) first.
        for (int r = 1; r <= available; r++) {
            mvwaddch(win, r, scroll_col, ' ');
        }
        // Draw the scrollbar using a block character.
        for (int r = 1; r <= scrollbar_height; r++) {
            mvwaddch(win, scrollbar_start + r, scroll_col, ACS_CKBOARD);
        }
    }
    wrefresh(win);
}


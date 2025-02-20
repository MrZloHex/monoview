#include "loger.h"
#include <string.h>


void
log_action(Loger *log, const char *action)
{
    if (log->q_entries < MAX_LOG_LINES)
    {
        strncpy(log->entry[log->q_entries], action, MAX_LOG_LENGTH - 1);
        log->entry[log->q_entries][MAX_LOG_LENGTH - 1] = '\0';
        log->q_entries++;
    }
    loger_update(log);
}

Loger
loger_init(int y, int x, int height, int width)
{
    Loger log;
    // TODO: check it
    
    log.focused   = false;
    log.q_entries = 0;
    log.scroll_offset = 0;

    log.win = newwin(height, width, y, x);
    wbkgd(log.win, COLOR_PAIR(1));
    wattron(log.win, COLOR_PAIR(1));
    box(log.win, 0, 0);
    wattroff(log.win, COLOR_PAIR(1));

    wrefresh(log.win);

    log.height = height;
    log.width  = width;

    return log;
}

#include "tui.h"

void
loger_update(Loger *loger)
{
    werase(loger->win);
    view_draw_focused((View)*loger);
    // The inner height available (excluding top and bottom borders)
    size_t available = loger->height - 2;
    size_t start_line = loger->scroll_offset;
    size_t end_line = (loger->q_entries < start_line + available) ? loger->q_entries : start_line + available;
    size_t line = 1;
    for (size_t i = start_line; i < end_line; i++, line++)
    {
        mvwprintw(loger->win, line, 1, "[%zu] %.*s", i, loger->width - 3, loger->entry[i]);
    }

    // Draw a scroll bar if necessary.
    if (loger->q_entries > available) {
        // Calculate the scrollbar height proportionally.
        size_t scrollbar_height = (available * available) / loger->q_entries;
        if (scrollbar_height < 1)
            scrollbar_height = 1;
        size_t max_offset = loger->q_entries - available;
        size_t scrollbar_start = 0;
        if (max_offset > 0)
            scrollbar_start = (loger->scroll_offset * (available - scrollbar_height)) / max_offset;
        // Choose the column for the scroll bar; here we use the second-to-last column (inside the right border).
        size_t scroll_col = loger->width - 2;
        // Clear that column (within inner area) first.
        for (size_t r = 1; r <= available; r++) {
            mvwaddch(loger->win, r, scroll_col, ' ');
        }
        // Draw the scrollbar using a block character.
        for (size_t r = 1; r <= scrollbar_height; r++) {
            mvwaddch(loger->win, scrollbar_start + r, scroll_col, ACS_CKBOARD);
        }
    }
    wrefresh(loger->win);
}

void
loger_pressed(Loger *loger, int ch)
{
    if (loger->q_entries <= loger->height -2)
    { return; }

    if (ch == KEY_UP)
    {
        if (loger->scroll_offset > 0)
        {
            loger->scroll_offset--;
            loger_update(loger);
        }
    }
    else if (ch == KEY_DOWN)
    {
        if (loger->scroll_offset < loger->q_entries - (loger->height - 2))
        {
            loger->scroll_offset++;
            loger_update(loger);
        }
    }
}


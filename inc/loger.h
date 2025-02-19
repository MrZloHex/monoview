#ifndef __LOGER_H__
#define __LOGER_H__

#include <ncurses.h>

#define MAX_LOG_LINES 100
#define MAX_LOG_LENGTH 256

extern char log_lines[MAX_LOG_LINES][MAX_LOG_LENGTH];
extern int num_log_lines;
extern int log_scroll_offset;

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;

    char entry[MAX_LOG_LINES][MAX_LOG_LENGTH];
    size_t q_entries;
} Loger;


Loger
loger_init(int y, int x, int height, int width);

void log_action(const char *action);

WINDOW *create_logs_window(int starty, int startx, int height, int width);

void update_logs_window(WINDOW *win, int height, int width);

#endif /* __LOGER_H__ */

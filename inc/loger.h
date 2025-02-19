#ifndef __LOGER_H__
#define __LOGER_H__

#include <ncurses.h>

#define MAX_LOG_LINES 100
#define MAX_LOG_LENGTH 256

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;

    char entry[MAX_LOG_LINES][MAX_LOG_LENGTH];
    size_t q_entries;
    size_t scroll_offset;
} Loger;


Loger
loger_init(int y, int x, int height, int width);

void
loger_update(Loger *loger);


void
log_action(Loger *log, const char *action);


#endif /* __LOGER_H__ */

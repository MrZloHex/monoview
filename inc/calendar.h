#ifndef __CALENDAR_H__
#define __CALENDAR_H__

#include <ncurses.h>

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;
} Calendar;

Calendar
calendar_init(int starty, int startx, int height, int width);

#endif /* __CALENDAR_H__ */

#ifndef __SCREENVIEW_H__
#define __SCREENVIEW_H__

#include <ncurses.h>

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;
} Screen;

Screen
screen_init(int x, int y, int height, int width);

void
screenview_update_datetime(WINDOW *win);


#endif /* __SCREENVIEW_H__ */

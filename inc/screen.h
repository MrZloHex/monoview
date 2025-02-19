#ifndef __SCREEN_H__
#define __SCREEN_H__

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
screen_update_datetime(Screen *scr);


#endif /* __SCREEN_H__ */

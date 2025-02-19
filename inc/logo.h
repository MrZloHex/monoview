#ifndef __LOGOVIEW_H__
#define __LOGOVIEW_H__

#include <ncurses.h>

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;
} Logo;

Logo
logo_init(int y, int x, int height, int width);

#endif /* __LOGOVIEW_H__ */


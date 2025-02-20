#ifndef __LOGO_H__
#define __LOGO_H__

#include <ncurses.h>

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;
} Logo;

Logo
logo_init(int y, int x, int height, int width);

#endif /* __LOGO_H__ */


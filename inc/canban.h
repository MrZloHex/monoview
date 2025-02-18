#ifndef __CANBAN_H__
#define __CANBAN_H__

#include "setup.h"

WINDOW *create_diary_window(int starty, int startx, int height, int width);
void update_diary_window(WINDOW *win, int height, int width);

#endif /* __CANBAN_H__ */

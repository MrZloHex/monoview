#ifndef TUI_H
#define TUI_H

#include "canban.h"
#include "emulator.h"
#include "logview.h"
#include "weekview.h"

void init_colors();
WINDOW *create_logo_window(int starty, int startx, int height, int width);
void draw_highlight(WINDOW *win, int focused);

#endif


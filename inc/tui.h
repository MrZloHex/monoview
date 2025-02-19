#ifndef TUI_H
#define TUI_H

#include "kanban.h"
#include "logview.h"
#include "weekview.h"

void init_colors();
void draw_highlight(WINDOW *win, int focused);

#endif


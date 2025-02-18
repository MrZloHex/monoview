#ifndef __EMULATOR_H__
#define __EMULATOR_H__

#include "setup.h"

WINDOW *create_lcd_window(int starty, int startx);
void update_lcd_time(WINDOW *lcd_win);

#endif /* __EMULATOR_H__ */

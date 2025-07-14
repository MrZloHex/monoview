#ifndef __UI_STATUS_H__
#define __UI_STATUS_H__

#include <notcurses/notcurses.h>

typedef struct
{
    struct ncplane *pl;
    struct ncplane *status;

    bool in_cmd_mode;
} StatusBar;

void
sb_init(StatusBar *sb, struct notcurses *nc);

void
sb_deinit(StatusBar *sb);

void
sb_enable_cmd(StatusBar *sb);

void
sb_disable_cmd(StatusBar *sb);

#endif /* __UI_STATUS_H__ */

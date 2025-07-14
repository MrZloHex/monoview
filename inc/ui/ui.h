#ifndef __UI_H__
#define __UI_H__

#include <notcurses/notcurses.h>

#include "ui/status.h"
#include "ui/cmd.h"
#include "ui/calendar.h"
#include "ui/notes.h"

typedef struct UI
{
    struct notcurses *nc;
    enum
    {
        MOD_REG,
        MOD_CMD,

        MOD_Q
    } mod;
    StatusBar         sb;
    CMD               cmd;

    bool should_close;
} UI;

void *
ui_thread(void *q);

#endif /* __UI_H__ */

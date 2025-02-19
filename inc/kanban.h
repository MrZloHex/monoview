#ifndef __CANBAN_H__
#define __CANBAN_H__

#include <ncurses.h>
#include "dynarray.h"

typedef struct {
    char name[256];
    char description[1024];
    char deadline[64];
    char label[64];
} KB_Card;

DECLARE_DYNARRAY(kb_bin, KB_Bin, KB_Card)

typedef struct
{
    KB_Bin todo;
    KB_Bin wip;
    KB_Bin done;

    WINDOW *win;
} Kanban;

Kanban *
kanban_init(int x, int y, int height, int width);

void
kanban_update(WINDOW *win, int height, int width);


#endif /* __CANBAN_H__ */

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
    WINDOW *win;
    bool focused;
    int height, width;

    KB_Bin todo;
    KB_Bin wip;
    KB_Bin done;
} Kanban;

Kanban
kanban_init(int x, int y, int height, int width);

void
kanban_update(Kanban *kan);


#endif /* __CANBAN_H__ */

#ifndef __CANBAN_H__
#define __CANBAN_H__

#include <ncurses.h>
#include "dynarray.h"

typedef struct {
    char name[256];
    char description[1024];
    time_t deadline;
    char label[64];
} KB_Card;

DECLARE_DYNARRAY(kb_vec, KB_Vec, KB_Card)

typedef struct
{
    KB_Vec cards;
    size_t card_focus;
    size_t width;

    size_t start, end;
} KB_Bin;

typedef enum
{
    BIN_TODO,
    BIN_WIP,
    BIN_DONE,
    BIN_QUANTITY
} KB_Bin_Types;

extern const char *kb_bin_names[BIN_QUANTITY];

typedef struct
{
    WINDOW *win;
    bool focused;
    int height, width;

    KB_Bin bins[BIN_QUANTITY];
    size_t bin_focus;
} Kanban;

Kanban
kanban_init(int x, int y, int height, int width);

void
kanban_update(Kanban *kan);

void
kanban_pressed(Kanban *kan, int ch);

void
kanban_add_entry(Kanban *kan);


#endif /* __CANBAN_H__ */

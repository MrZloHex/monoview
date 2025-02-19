#include "kanban.h"



Kanban
kanban_init(int x, int y, int height, int width)
{
    Kanban kan;
    kb_bin_init(&kan.todo, 8);
    kb_bin_init(&kan.wip,  8);
    kb_bin_init(&kan.done, 8);
    // TODO: add check

    kan.height = height;
    kan.width  = width;
    kan.win = newwin(height, width, x, y);
    wbkgd(kan.win, COLOR_PAIR(1));

    kb_bin_append(&kan.todo, (KB_Card){ .name = "HUY" });

    wrefresh(kan.win);
    return kan;
}



void
kanban_update(Kanban *kan)
{
    werase(kan->win);
    int col_width = (kan->width - 2) / 3;

    mvwprintw(kan->win, 1, 1 + (col_width - 6) / 2, "To Do");
    mvwprintw(kan->win, 1, 1 + col_width + (col_width - 10) / 2, "In Prog");
    mvwprintw(kan->win, 1, 1 + 2 * col_width + (col_width - 4) / 2, "Done");
    int line = 3;
    for (size_t i = 0; i < kan->todo.size && line < kan->height - 1; i++, line++) {
        char display[300];
        snprintf(display, sizeof(display), "- %s (%s)", kan->todo.data[i].name, kan->todo.data[i].deadline);
        mvwprintw(kan->win, line, 1, "%s", display);
    }
    wrefresh(kan->win);
}

DEFINE_DYNARRAY(kb_bin, KB_Bin, KB_Card)

#include "kanban.h"



Kanban *
kanban_init(int x, int y, int height, int width)
{
    Kanban *kan = (Kanban *)malloc(sizeof(Kanban));
    // TODO: add check

    kb_bin_init(&kan->todo, 8);
    kb_bin_init(&kan->wip,  8);
    kb_bin_init(&kan->done, 8);

    kan->win = newwin(height, width, x, y);
    wbkgd(kan->win, COLOR_PAIR(1));

    wrefresh(kan->win);
    return kan;
}



void
kanban_update(WINDOW *win, int height, int width)
{
    werase(win);
    wattron(win, COLOR_PAIR(1));
    box(win, 0, 0);
    wattroff(win, COLOR_PAIR(1));
    int col_width = (width - 2) / 3;
    mvwprintw(win, 1, 1 + (col_width - 6) / 2, "To Do");
    mvwprintw(win, 1, 1 + col_width + (col_width - 10) / 2, "In Prog");
    mvwprintw(win, 1, 1 + 2 * col_width + (col_width - 4) / 2, "Done");
    int line = 3;
    // for (int i = 0; i < num_entries && line < height - 1; i++, line++) {
    //     char display[300];
    //     snprintf(display, sizeof(display), "- %s (%s)", kanban_entries[i].name, kanban_entries[i].deadline);
    //     mvwprintw(win, line, 1, "%s", display);
    // }
    wrefresh(win);
}

DEFINE_DYNARRAY(kb_bin, KB_Bin, KB_Card)

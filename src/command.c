#include "command.h"
#include "setup.h"
#include "tui.h"
#include <ncurses.h>
#include <string.h>
#include <unistd.h>


// Global arrays declared in global.h are used here.
extern int num_entries;
extern int command_quit;

#if 0
void enter_command_mode(int max_y, int max_x, WINDOW *vert_sep_top, WINDOW *vert_sep_bottom,
                          WINDOW *horz_sep, WINDOW **win_array, int num_windows,
                          WINDOW *diary_win, int diary_height, int diary_width,
                          WINDOW *lcd_win, WINDOW *logo_win) {
    int cmd_height = 3;
    WINDOW *cmd_win = newwin(cmd_height, max_x, max_y - cmd_height, 0);
    wbkgd(cmd_win, COLOR_PAIR(1));
    box(cmd_win, 0, 0);
    mvwprintw(cmd_win, 1, 1, "CMD: ");
    wrefresh(cmd_win);
    nodelay(cmd_win, TRUE);

    char cmd[80] = {0};
    int idx = 0;
    int ch;

    while (1) {
        update_lcd_time(lcd_win);
        ch = wgetch(cmd_win);
        if (ch != ERR) {
            if (ch == '\n' || ch == KEY_ENTER) break;
            else if (ch == KEY_BACKSPACE || ch == 127 || ch == '\b') {
                if (idx > 0) { idx--; cmd[idx] = '\0'; }
            } else {
                if (idx < 79) { cmd[idx++] = ch; cmd[idx] = '\0'; }
            }
            werase(cmd_win);
            box(cmd_win, 0, 0);
            mvwprintw(cmd_win, 1, 1, "CMD: %s", cmd);
            wrefresh(cmd_win);
        }
        napms(100);
    }

    if (strncmp(cmd, "add ", 4) == 0) {
        char *task = cmd + 4;
        // For backward compatibility, we could add to a simple list.
        // Here we add a new entry with only name and deadline.
        if (num_entries < MAX_ENTRIES) {
            strncpy(kanban_entries[num_entries].name, task, 255);
            kanban_entries[num_entries].name[255] = '\0';
            strncpy(kanban_entries[num_entries].deadline, "N/A", 63);
            kanban_entries[num_entries].deadline[63] = '\0';
            // Other fields remain empty.
            num_entries++;
        }
    }
    if (strncmp(cmd, "quit", 4) == 0) {
        command_quit = 1;
    }

    werase(cmd_win);
    wrefresh(cmd_win);
    delwin(cmd_win);

    clear();
    refresh();
    wrefresh(vert_sep_top);
    wrefresh(vert_sep_bottom);
    wrefresh(horz_sep);
    for (int i = 0; i < num_windows; i++) {
        touchwin(win_array[i]);
        wrefresh(win_array[i]);
    }
    touchwin(logo_win);
    wrefresh(logo_win);
    update_diary_window(diary_win, diary_height, diary_width);
}


#endif

void enter_new_entry(WINDOW *diary_win, int diary_height, int diary_width) {
    int dlg_height = 12;
    int dlg_width = 60;
    int starty = (LINES - dlg_height) / 2;
    int startx = (COLS - dlg_width) / 2;
    WINDOW *dlg = newwin(dlg_height, dlg_width, starty, startx);
    wbkgd(dlg, COLOR_PAIR(1));
    box(dlg, 0, 0);
    mvwprintw(dlg, 1, 2, "New Kanban Entry");
    wrefresh(dlg);

    char name[256] = {0};
    char description[1024] = {0};
    char deadline[64] = {0};
    char label[64] = {0};

    // Enable echo so input is visible.
    echo();

    mvwprintw(dlg, 2, 2, "Name: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 2, 8, name, 255);

    mvwprintw(dlg, 3, 2, "Description: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 3, 15, description, 1023);

    mvwprintw(dlg, 4, 2, "Deadline: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 4, 12, deadline, 63);

    mvwprintw(dlg, 5, 2, "Label: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 5, 10, label, 63);

    // Disable echo after input is done.
    noecho();

    // Now store the entry...
    if (num_entries < MAX_ENTRIES) {
        strncpy(kanban_entries[num_entries].name, name, 255);
        kanban_entries[num_entries].name[255] = '\0';
        strncpy(kanban_entries[num_entries].description, description, 1023);
        kanban_entries[num_entries].description[1023] = '\0';
        strncpy(kanban_entries[num_entries].deadline, deadline, 63);
        kanban_entries[num_entries].deadline[63] = '\0';
        strncpy(kanban_entries[num_entries].label, label, 63);
        kanban_entries[num_entries].label[63] = '\0';
        num_entries++;
    }

    werase(dlg);
    wrefresh(dlg);
    delwin(dlg);

    update_diary_window(diary_win, diary_height, diary_width);
}


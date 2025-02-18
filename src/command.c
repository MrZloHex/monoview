#include "command.h"
#include "setup.h"
#include "tui.h"
#include <ncurses.h>
#include <string.h>
#include <unistd.h>

extern char todo_list[MAX_TODO][256];
extern int todo_count;
extern int command_quit;

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
    
    // Set non-blocking mode.
    nodelay(cmd_win, TRUE);
    char cmd[80];
    int idx = 0;
    memset(cmd, 0, sizeof(cmd));
    int ch;
    
    while (1) {
        update_lcd_time(lcd_win);
        ch = wgetch(cmd_win);
        if (ch != ERR) {
            if (ch == '\n' || ch == KEY_ENTER) {
                break;
            } else if (ch == KEY_BACKSPACE || ch == 127 || ch == '\b') {
                if (idx > 0) {
                    idx--;
                    cmd[idx] = '\0';
                }
            } else {
                if (idx < 79) {
                    cmd[idx++] = ch;
                    cmd[idx] = '\0';
                }
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
        if (todo_count < MAX_TODO) {
            strncpy(todo_list[todo_count], task, 255);
            todo_list[todo_count][255] = '\0';
            todo_count++;
        }
    }
    if (strncmp(cmd, "quit", 4) == 0) {
        command_quit = 1;
    }
    
    werase(cmd_win);
    wrefresh(cmd_win);
    delwin(cmd_win);
    
    // Redraw layout.
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


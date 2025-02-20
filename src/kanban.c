#include "kanban.h"

#include <string.h>
#include <time.h>



Kanban
kanban_init(int x, int y, int height, int width)
{
    Kanban kan;
    // TODO: add check

    kan.height    = height;
    kan.width     = width;
    kan.focused   = false;
    kan.bin_focus = 0;
    kan.win = newwin(height, width, x, y);
    wbkgd(kan.win, COLOR_PAIR(1));

    size_t bin_width = ((size_t)width - BIN_QUANTITY -1) / BIN_QUANTITY;
    size_t error = ((size_t)width - BIN_QUANTITY -1) % BIN_QUANTITY;
    bin_width += error / BIN_QUANTITY;
    error = ((size_t)width - BIN_QUANTITY -1)- bin_width*BIN_QUANTITY;

    // mvwprintw(kan.win, 7, 1, "%zu %zu", bin_width, error);
    for (size_t i = 0; i < BIN_QUANTITY; ++i)
    {
        kb_vec_init(&kan.bins[i].cards, 8);
        kan.bins[i].card_focus = 0;
        kan.bins[i].width      = bin_width;
    }

    if (error % 2)
    {
        kan.bins[BIN_QUANTITY /2].width += error;
    }
    else
    {
        kan.bins[0].width               += error / 2;
        kan.bins[BIN_QUANTITY -1].width += error / 2;
    }

    // for (size_t i = 0; i < BIN_QUANTITY; ++i)
    // {
    //     mvwprintw(kan.win, 8+i, 1, "%zu %zu", i, kan.bins[i].width); 
    // }


    wrefresh(kan.win);
    return kan;
}


void draw_card(WINDOW *win, int y, int x, int col_width, KB_Card card) {
    // First line: name and label.
    wattron(win, COLOR_PAIR(2));

    mvwprintw(win, y, x, "%-20.20s", card.name);
    mvwprintw(win, y, x + col_width - 10, "%10.10s", card.label);
    wattroff(win, COLOR_PAIR(2));
    // Second line: short description.
    mvwprintw(win, y + 1, x + 3, "%.30s", card.description);
    // Third line: deadline formatted as "YYYY-MM-DD HH:MM", right aligned.
    char deadline_str[32];
    strftime(deadline_str, sizeof(deadline_str), "%Y-%m-%d %H:%M", localtime(&card.deadline));
    mvwprintw(win, y + 2, x + col_width - 16, "%15s", deadline_str);
}

#include "tui.h"

void
kanban_update(Kanban *kan)
{
    werase(kan->win);
    view_draw_focused((View)*kan);

    for (int y = 1; y < kan->height - 1; ++y)
    {
        size_t sum_width = 1 + kan->bins[0].width;
        for (size_t i = 1; i < BIN_QUANTITY; ++i)
        {
            mvwaddch(kan->win, y, sum_width, ACS_VLINE);
            sum_width += 1 + kan->bins[i].width;
        }
    }

    int next_y[BIN_QUANTITY] = { 3, 3, 3 };

    size_t sum_width = 1;
    for (size_t bi = 0; bi < BIN_QUANTITY; ++bi)
    {
        KB_Bin bin = kan->bins[bi];
        // HEADER
        wattron(kan->win, COLOR_PAIR(bi == kan->bin_focus ? 2 : 1));
        size_t name_len = strlen(kb_bin_names[bi]);
        mvwprintw(kan->win, 1, sum_width + (bin.width - name_len)/2, kb_bin_names[bi]);
        wattroff(kan->win, COLOR_PAIR(bi == kan->bin_focus ? 2 : 1));

        for (size_t i = 0; i < bin.cards.size; ++i)
        {
            if (next_y[bi] + 3 > kan->height - 1)
            { break; }
            
            if (i == bin.card_focus && bi == kan->bin_focus)
            {
                mvwhline(kan->win, next_y[bi]-1, sum_width, 0, bin.width);
            }

            KB_Card card;
            kb_vec_get(&bin.cards, i, &card);
            draw_card(kan->win, next_y[bi], sum_width, bin.width, card);
            next_y[bi] += 4;

            if (i == bin.card_focus && bi == kan->bin_focus)
            {
                mvwhline(kan->win, next_y[bi]-1, sum_width, 0, bin.width);
            }
        }

        sum_width += 1 + bin.width;
    }


    wrefresh(kan->win);

}

void
kanban_pressed(Kanban *kan, int ch)
{
    if (ch == 'i')
    {
        kanban_add_entry(kan);
    }
    if (ch == KEY_LEFT)
    {
        kan->bin_focus = (kan->bin_focus +5) % BIN_QUANTITY;
    }
    else if (ch == KEY_RIGHT)
    {
        kan->bin_focus = (kan->bin_focus +1) % BIN_QUANTITY;
    }
    else if (ch == KEY_UP)
    {
        KB_Bin *bin = &kan->bins[kan->bin_focus];
        if (bin->cards.size > 1)
        {
            bin->card_focus -= 1;
            bin->card_focus %= bin->cards.size;
        }
    }
    else if (ch == KEY_DOWN)
    {
        KB_Bin *bin = &kan->bins[kan->bin_focus];
        if (bin->cards.size > 1)
        {
            bin->card_focus += 1;
            bin->card_focus %= bin->cards.size;
        }
    }
    kanban_update(kan);
}


void
draw_datetime(WINDOW *win, int row, int col, struct tm *tm_val, int active_field)
{
    char year_str[5], mon_str[3], day_str[3], hour_str[3], min_str[3];
    snprintf(year_str, sizeof(year_str), "%04d", tm_val->tm_year + 1900);
    snprintf(mon_str, sizeof(mon_str), "%02d", tm_val->tm_mon + 1);
    snprintf(day_str, sizeof(day_str), "%02d", tm_val->tm_mday);
    snprintf(hour_str, sizeof(hour_str), "%02d", tm_val->tm_hour);
    snprintf(min_str, sizeof(min_str), "%02d", tm_val->tm_min);

    // Print the fixed prefix.
    mvwprintw(win, row, col, "Deadline: ");
    int pos = col + 10; // "Deadline: " is 10 characters

    // Print year
    if (active_field == 0) wattron(win, A_REVERSE);
    mvwprintw(win, row, pos, "%s", year_str);
    if (active_field == 0) wattroff(win, A_REVERSE);
    pos += 4;
    mvwprintw(win, row, pos, "-");
    pos++;
    // Print month
    if (active_field == 1) wattron(win, A_REVERSE);
    mvwprintw(win, row, pos, "%s", mon_str);
    if (active_field == 1) wattroff(win, A_REVERSE);
    pos += 2;
    mvwprintw(win, row, pos, "-");
    pos++;
    // Print day
    if (active_field == 2) wattron(win, A_REVERSE);
    mvwprintw(win, row, pos, "%s", day_str);
    if (active_field == 2) wattroff(win, A_REVERSE);
    pos += 2;
    mvwprintw(win, row, pos, " ");
    pos++;
    // Print hour
    if (active_field == 3) wattron(win, A_REVERSE);
    mvwprintw(win, row, pos, "%s", hour_str);
    if (active_field == 3) wattroff(win, A_REVERSE);
    pos += 2;
    mvwprintw(win, row, pos, ":");
    pos++;
    // Print minute
    if (active_field == 4) wattron(win, A_REVERSE);
    mvwprintw(win, row, pos, "%s", min_str);
    if (active_field == 4) wattroff(win, A_REVERSE);

    wrefresh(win);
}

// Interactive datetime input using arrow keys to adjust fields.
// Returns the final time_t value.
time_t
input_datetime(WINDOW *win, int start_row, int start_col)
{
    // Get current time.
    time_t now = time(NULL);
    struct tm tm_val;
    localtime_r(&now, &tm_val);

    int field = 0;  // active field: 0=year,1=month,2=day,3=hour,4=minute
    bool done = false;

    // Ensure keypad is enabled.
    keypad(win, TRUE);
    // Show cursor if desired.
    curs_set(0);  // We'll highlight with reverse video instead.

    while (!done) {
        // Redraw the datetime with the active field highlighted.
        draw_datetime(win, start_row, start_col, &tm_val, field);
        int ch = wgetch(win);
        switch (ch) {
            case KEY_LEFT:
                if (field > 0) field--;
                break;
            case KEY_RIGHT:
                if (field < 4) field++;
                break;
            case KEY_UP:
                switch (field) {
                    case 0: tm_val.tm_year++; break;
                    case 1: tm_val.tm_mon = (tm_val.tm_mon + 1) % 12; break;
                    case 2: tm_val.tm_mday++; break;
                    case 3: tm_val.tm_hour = (tm_val.tm_hour + 1) % 24; break;
                    case 4: tm_val.tm_min = (tm_val.tm_min + 1) % 60; break;
                }
                break;
            case KEY_DOWN:
                switch (field) {
                    case 0: tm_val.tm_year--; break;
                    case 1: tm_val.tm_mon = (tm_val.tm_mon + 11) % 12; break;
                    case 2: tm_val.tm_mday = (tm_val.tm_mday > 1 ? tm_val.tm_mday - 1 : 1); break;
                    case 3: tm_val.tm_hour = (tm_val.tm_hour + 23) % 24; break;
                    case 4: tm_val.tm_min = (tm_val.tm_min + 59) % 60; break;
                }
                break;
            case '\n':
            case KEY_ENTER:
                done = true;
                break;
            default:
                break;
        }
        // Optionally, you could call mktime(&tm_val) here to normalize the structure.
    }
    return mktime(&tm_val);
}



void
kanban_add_entry(Kanban *kan)
{
    int dlg_height = 12;
    int dlg_width = 60;
    int starty = (LINES - dlg_height) / 2;
    int startx = (COLS - dlg_width) / 2;
    WINDOW *dlg = newwin(dlg_height, dlg_width, starty, startx);
    wbkgd(dlg, COLOR_PAIR(1));
    box(dlg, 0, 0);
    mvwprintw(dlg, 1, 2, "Add Card");
    wrefresh(dlg);

    char name[256] = {0};
    char description[1024] = {0};
    char label[64] = {0};

    echo();  // Enable character input display

    // Name input
    mvwprintw(dlg, 2, 2, "Name: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 2, 8, name, 255);

    // Description input
    mvwprintw(dlg, 3, 2, "Description: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 3, 15, description, 1023);

    // Interactive datetime input for Deadline
    mvwprintw(dlg, 4, 2, "Deadline: ");
    wrefresh(dlg);
    time_t deadline = input_datetime(dlg, 4, 12);

    // Label input
    mvwprintw(dlg, 5, 2, "Label: ");
    wrefresh(dlg);
    mvwgetnstr(dlg, 5, 10, label, 63);

    noecho();  // Disable character input display

    // Create and populate the new card
    KB_Card card;
    strncpy(card.name, name, sizeof(card.name) - 1);
    card.name[sizeof(card.name) - 1] = '\0';
    strncpy(card.description, description, sizeof(card.description) - 1);
    card.description[sizeof(card.description) - 1] = '\0';
    strncpy(card.label, label, sizeof(card.label) - 1);
    card.label[sizeof(card.label) - 1] = '\0';
    card.deadline = deadline;

    // Append the new card to the todo bin
    kb_vec_append(&kan->bins[kan->bin_focus].cards, card);

    // Don't use werase to clear, instead use wclear to clear content area
    wclear(dlg);  // This will clear the content but keep the borders
    wrefresh(dlg);  // Refresh the content after clearing

    delwin(dlg);  // Close the dialog window

    // Update the Kanban view with the new entry
}



DEFINE_DYNARRAY(kb_vec, KB_Vec, KB_Card)

const char *kb_bin_names[BIN_QUANTITY] =
{ "TODO", "WIP", "DONE" };

/* kanban.c */

#include "logic/kanban.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define INITIAL_CAP 4

/* Утилита для расширения динамического массива */
static void **resize_array(void **arr, size_t *cap) {
    size_t new_cap = (*cap == 0 ? INITIAL_CAP : *cap * 2);
    void **tmp = realloc(arr, new_cap * sizeof(void *));
    if (!tmp) {
        perror("realloc");
        exit(EXIT_FAILURE);
    }
    *cap = new_cap;
    return tmp;
}

/* === Board === */

Board *
board_create(void) {
    Board *b = calloc(1, sizeof(Board));
    if (!b) exit(EXIT_FAILURE);
    return b;
}

void
board_destroy(Board *b) {
    if (!b) return;
    for (size_t i = 0; i < b->list_count; ++i)
        list_destroy(b->lists[i]);
    free(b->lists);
    free(b);
}

List *
board_add_list(Board *b, const char *name) {
    if (b->list_count == b->list_capacity)
        b->lists = (List **)resize_array((void **)b->lists, &b->list_capacity);

    List *l = malloc(sizeof(List));
    l->name          = strdup(name);
    l->cards         = NULL;
    l->card_count    = 0;
    l->card_capacity = 0;

    b->lists[b->list_count++] = l;
    return l;
}

List *
board_find_list(Board *b, const char *name) {
    for (size_t i = 0; i < b->list_count; ++i) {
        if (strcmp(b->lists[i]->name, name) == 0)
            return b->lists[i];
    }
    return NULL;
}

int
board_move_card_to(Board *b,
                   int card_id,
                   const char *from,
                   const char *to,
                   size_t pos) {
    List *src = board_find_list(b, from);
    List *dst = board_find_list(b, to);
    if (!src || !dst) return -1;

    /* Найти карточку в src */
    Card *c = NULL;
    size_t idx = 0;
    for (; idx < src->card_count; ++idx) {
        if (src->cards[idx]->id == card_id) {
            c = src->cards[idx];
            break;
        }
    }
    if (!c) return -2;

    /* Удалить из src */
    memmove(&src->cards[idx],
            &src->cards[idx + 1],
            (src->card_count - idx - 1) * sizeof(Card *));
    --src->card_count;

    /* Вставить в dst на позицию pos */
    if (dst->card_count == dst->card_capacity)
        dst->cards = (Card **)resize_array((void **)dst->cards,
                                           &dst->card_capacity);

    if (pos > dst->card_count) pos = dst->card_count;
    memmove(&dst->cards[pos + 1],
            &dst->cards[pos],
            (dst->card_count - pos) * sizeof(Card *));
    dst->cards[pos] = c;
    ++dst->card_count;

    return 0;
}

void
board_display(const Board *b) {
    printf("=== BOARD ===\n");
    for (size_t i = 0; i < b->list_count; ++i) {
        List *l = b->lists[i];
        printf("-- %s [%zu cards] --\n", l->name, l->card_count);
        for (size_t j = 0; j < l->card_count; ++j) {
            Card *c = l->cards[j];
            char buf_created[20], buf_deadline[20];
            struct tm tm;

            localtime_r(&c->created, &tm);
            strftime(buf_created, sizeof(buf_created), "%Y-%m-%d", &tm);
            localtime_r(&c->deadline, &tm);
            strftime(buf_deadline, sizeof(buf_deadline), "%Y-%m-%d", &tm);

            const char *color_str;
            switch (c->color) {
                case KB_COLOR_RED:    color_str = "RED";    break;
                case KB_COLOR_GREEN:  color_str = "GREEN";  break;
                case KB_COLOR_BLUE:   color_str = "BLUE";   break;
                case KB_COLOR_YELLOW: color_str = "YELLOW"; break;
                default:           color_str = "NONE";   break;
            }

            printf(" [%d] %s\n"
                   "     desc    : %s\n"
                   "     created : %s\n"
                   "     deadline: %s\n"
                   "     color   : %s\n",
                   c->id, c->title,
                   c->desc,
                   buf_created,
                   buf_deadline,
                   color_str);

            if (c->item_count) {
                printf("     checklist:\n");
                for (size_t k = 0; k < c->item_count; ++k) {
                    ChecklistItem *it = c->items[k];
                    printf("       - [%c] %s\n",
                           it->done ? 'x' : ' ',
                           it->text);
                }
            }
        }
    }
}

/* === List & Card helpers === */

Card *
list_add_card(List *l,
              int id,
              const char *title,
              const char *desc,
              time_t deadline,
              CardColor color) {
    if (l->card_count == l->card_capacity)
        l->cards = (Card **)resize_array((void **)l->cards,
                                         &l->card_capacity);

    Card *c = malloc(sizeof(Card));
    c->id            = id;
    c->title         = strdup(title);
    c->desc          = strdup(desc);
    c->created       = time(NULL);
    c->deadline      = deadline;
    c->color         = color;
    c->items         = NULL;
    c->item_count    = 0;
    c->item_capacity = 0;

    l->cards[l->card_count++] = c;
    return c;
}

int
list_move_card(List *l, size_t from, size_t to) {
    if (from >= l->card_count || to > l->card_count) return -1;
    Card *c = l->cards[from];
    if (from < to) {
        memmove(&l->cards[from],
                &l->cards[from + 1],
                (to - from) * sizeof(Card *));
    } else if (from > to) {
        memmove(&l->cards[to + 1],
                &l->cards[to],
                (from - to) * sizeof(Card *));
    }
    l->cards[to] = c;
    return 0;
}

void
list_destroy(List *l) {
    if (!l) return;
    for (size_t i = 0; i < l->card_count; ++i) {
        Card *c = l->cards[i];
        free(c->title);
        free(c->desc);
        for (size_t j = 0; j < c->item_count; ++j) {
            free(c->items[j]->text);
            free(c->items[j]);
        }
        free(c->items);
        free(c);
    }
    free(l->cards);
    free(l->name);
    free(l);
}

int
card_add_checklist_item(Card *c, int item_id, const char *text) {
    if (c->item_count == c->item_capacity)
        c->items = (ChecklistItem **)resize_array((void **)c->items,
                                                  &c->item_capacity);

    ChecklistItem *it = malloc(sizeof(ChecklistItem));
    it->id   = item_id;
    it->text = strdup(text);
    it->done = 0;
    c->items[c->item_count++] = it;
    return 0;
}

int
card_toggle_checklist_item(Card *c, int item_id) {
    for (size_t i = 0; i < c->item_count; ++i) {
        if (c->items[i]->id == item_id) {
            c->items[i]->done = !c->items[i]->done;
            return 0;
        }
    }
    return -1;
}


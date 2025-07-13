#ifndef KANBAN_H
#define KANBAN_H

#include <stddef.h>
#include <time.h>


/* Цвет карточки */
typedef enum {
    KB_COLOR_NONE = 0,
    KB_COLOR_RED,
    KB_COLOR_GREEN,
    KB_COLOR_BLUE,
    KB_COLOR_YELLOW
} CardColor;

/* Элемент чек-листа */
typedef struct ChecklistItem {
    int     id;
    char   *text;
    int     done;
} ChecklistItem;

/* Карточка */
typedef struct Card {
    int               id;
    char             *title;
    char             *desc;
    time_t            created;
    time_t            deadline;
    CardColor         color;
    ChecklistItem   **items;
    size_t            item_count;
    size_t            item_capacity;
} Card;

/* Список (колонка) */
typedef struct List {
    char   *name;
    Card  **cards;
    size_t  card_count;
    size_t  card_capacity;
} List;

/* Доска */
typedef struct Board {
    List  **lists;
    size_t  list_count;
    size_t  list_capacity;
} Board;

/* === Board API === */
Board *   board_create(void);
void      board_destroy(Board *b);
List *    board_add_list(Board *b, const char *name);
List *    board_find_list(Board *b, const char *name);
int       board_move_card_to(Board *b,
                             int card_id,
                             const char *from,
                             const char *to,
                             size_t pos);

/* === List API === */
Card *    list_add_card(List *l,
                       int id,
                       const char *title,
                       const char *desc,
                       time_t deadline,
                       CardColor color);
int       list_move_card(List *l, size_t from_pos, size_t to_pos);
void      list_destroy(List *l);

/* === Card API === */
int       card_add_checklist_item(Card *c, int item_id, const char *text);
int       card_toggle_checklist_item(Card *c, int item_id);

/* === Display === */
void      board_display(const Board *b);

#endif /* KANBAN_H */

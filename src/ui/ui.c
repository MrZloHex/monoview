/*
 * module_ui.c
 * A generic ncurses UI framework without globals.
 * Static 2x2 layout (no auto-resize).
 * Entry point: void module_ui_entry(void *arg)
 */

#define _POSIX_C_SOURCE 200809L
#include <ncurses.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdio.h>

// Abstract Module structure
typedef struct Module {
    WINDOW *win;
    bool active;
    void *data;
    void (*draw)(struct Module *self);
    void (*handle_input)(struct Module *self, int ch);
} Module;

// UI Context holding modules and state
typedef struct {
    Module **modules;
    int module_count;
    int current_index;
    int max_modules;
    int height, width;
    WINDOW *cmd_win;
    bool cmd_mode;
    char cmd_buf[256];
    int cmd_pos;
} UIContext;

// Label module data
typedef struct { char *label; } LabelData;

// Forward declarations
static UIContext* ui_context_create(int max_modules);
static void ui_context_destroy(UIContext *ctx);
static int ui_add_module(UIContext *ctx, Module *mod);
static void draw_cmd(UIContext *ctx);
static void label_draw(Module *self);
static void label_handle(Module *self, int ch);
static Module* create_label_module(const char *text, int y, int x, int h, int w);

// Create UI context
static UIContext* ui_context_create(int max_modules) {
    UIContext *ctx = malloc(sizeof(UIContext));
    ctx->modules = calloc(max_modules, sizeof(Module*));
    ctx->module_count = 0;
    ctx->current_index = 0;
    ctx->max_modules = max_modules;
    ctx->height = ctx->width = 0;
    ctx->cmd_mode = false;
    ctx->cmd_pos = 0;
    memset(ctx->cmd_buf, 0, sizeof(ctx->cmd_buf));
    return ctx;
}

static void ui_context_destroy(UIContext *ctx) {
    free(ctx->modules);
    free(ctx);
}

static int ui_add_module(UIContext *ctx, Module *mod) {
    if (ctx->module_count >= ctx->max_modules) return -1;
    ctx->modules[ctx->module_count++] = mod;
    return ctx->module_count - 1;
}

// Draw bottom command/status bar
static void draw_cmd(UIContext *ctx) {
    werase(ctx->cmd_win);
    if (ctx->cmd_mode) {
        wattron(ctx->cmd_win, COLOR_PAIR(3));
        box(ctx->cmd_win, 0, 0);
        wattroff(ctx->cmd_win, COLOR_PAIR(3));
        mvwprintw(ctx->cmd_win, 1, 1, ":%s", ctx->cmd_buf);
    } else {
        box(ctx->cmd_win, 0, 0);
        mvwprintw(ctx->cmd_win, 1, 1, "Press ':' for commands.");
    }
    wrefresh(ctx->cmd_win);
}

// Label module functions
static void label_draw(Module *self) {
    werase(self->win);
    if (self->active) wattron(self->win, COLOR_PAIR(2));
    box(self->win, 0, 0);
    if (self->active) wattroff(self->win, COLOR_PAIR(2));
    LabelData *d = (LabelData*)self->data;
    mvwprintw(self->win, 1, 2, "%s", d->label);
    wrefresh(self->win);
}

static void label_handle(Module *self, int ch) {
    (void)self; (void)ch;
}

static Module* create_label_module(const char *text, int y, int x, int h, int w) {
    Module *m = malloc(sizeof(Module));
    m->win = newwin(h, w, y, x);
    m->active = false;
    LabelData *d = malloc(sizeof(LabelData));
    d->label = strdup(text);
    m->data = d;
    m->draw = label_draw;
    m->handle_input = label_handle;
    return m;
}

// Entry point: sets up context, static modules, and runs loop
void module_ui_entry(void *arg) {
    (void)arg;
    int rows = 2, cols = 2;
    int maxm = rows * cols;
    UIContext *ctx = ui_context_create(maxm);

    initscr();
    set_escdelay(25);
    cbreak();
    noecho();
    keypad(stdscr, TRUE);
    if (has_colors()) {
        start_color();
        init_pair(1, COLOR_WHITE, COLOR_BLACK);
        init_pair(2, COLOR_YELLOW, COLOR_BLACK);
        init_pair(3, COLOR_CYAN, COLOR_BLACK);
    }
    clear(); refresh();

    getmaxyx(stdscr, ctx->height, ctx->width);
    ctx->cmd_win = newwin(3, ctx->width, ctx->height - 3, 0);

    int mod_h = (ctx->height - 3) / rows;
    int mod_w = ctx->width / cols;
    for (int i = 0; i < maxm; i++) {
        int r = i / cols;
        int c = i % cols;
        char buf[32];
        snprintf(buf, sizeof(buf), "Module %d", i + 1);
        Module *m = create_label_module(buf, r * mod_h, c * mod_w, mod_h, mod_w);
        ui_add_module(ctx, m);
    }

    for (int i = 0; i < ctx->module_count; i++) {
        Module *m = ctx->modules[i];
        m->active = (i == ctx->current_index);
        m->draw(m);
    }
    draw_cmd(ctx);
    refresh();

    int ch;
    while ((ch = getch()) != ERR) {
        if (ctx->cmd_mode) {
            if (ch == 27) {
                ctx->cmd_mode = false;
                ctx->cmd_pos = 0;
                ctx->cmd_buf[0] = '\0';
                draw_cmd(ctx);
                continue;
            }
            if (ch == '\n') {
                if (strcmp(ctx->cmd_buf, "q") == 0) break;
                ctx->cmd_mode = false;
                ctx->cmd_pos = 0;
                ctx->cmd_buf[0] = '\0';
                draw_cmd(ctx);
                continue;
            }
            if (ch == KEY_BACKSPACE || ch == 127) {
                if (ctx->cmd_pos > 0) ctx->cmd_buf[--ctx->cmd_pos] = '\0';
                draw_cmd(ctx);
                continue;
            }
            if (ch >= 32 && ch <= 126 && ctx->cmd_pos < (int)sizeof(ctx->cmd_buf) - 1) {
                ctx->cmd_buf[ctx->cmd_pos++] = (char)ch;
                ctx->cmd_buf[ctx->cmd_pos] = '\0';
                draw_cmd(ctx);
            }
        } else {
            if (ch == '\t') {
                Module *old = ctx->modules[ctx->current_index];
                old->active = false;
                old->draw(old);
                ctx->current_index = (ctx->current_index + 1) % ctx->module_count;
                Module *nw = ctx->modules[ctx->current_index];
                nw->active = true;
                nw->draw(nw);
                continue;
            }
            if (ch == ':') {
                ctx->cmd_mode = true;
                ctx->cmd_pos = 0;
                ctx->cmd_buf[0] = '\0';
                draw_cmd(ctx);
                continue;
            }
        }
    }

    for (int i = 0; i < ctx->module_count; i++) {
        Module *m = ctx->modules[i];
        LabelData *d = (LabelData*)m->data;
        free(d->label);
        free(d);
        delwin(m->win);
        free(m);
    }
    delwin(ctx->cmd_win);
    endwin();
    ui_context_destroy(ctx);
}


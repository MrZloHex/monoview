#define _XOPEN_SOURCE 700          /* даёт объявления wcwidth() и wcswidth() */
#include <locale.h>
#include <stdlib.h>
#include <notcurses/notcurses.h>

static void
tab1_cb(struct nctab* t, struct ncplane* ncp, void* curry){
    ncplane_printf(ncp, "Это вкладка 1\n");
}

static void
tab2_cb(struct nctab* t, struct ncplane* ncp, void* curry){
    ncplane_printf(ncp, "Это вкладка 2\n");
}

int main(void){
    setlocale(LC_ALL, "");
    struct notcurses* nc = notcurses_init(NULL, NULL);
    if(!nc) return EXIT_FAILURE;

    unsigned rows, cols;
    notcurses_stddim_yx(nc, &rows, &cols);              // получить размеры терминала :contentReference[oaicite:0]{index=0}
    struct ncplane* stdp = notcurses_stdplane(nc);      // стандартная плоскость

    // создаём «фрейм» для вкладок, например со смещением (1,2) и отступом по краям
    struct ncplane_options popts = {
        .y     = 1,
        .x     = 2,
        .rows  = rows - 4,
        .cols  = cols - 4,
        .flags = 0,
    };
    struct ncplane* host = ncplane_create(stdp, &popts);

    // опции виджета вкладок: цвета заголовков, выбранной вкладки и разделителя
    struct nctabbed_options topts = {
        .hdrchan   = NCCHANNELS_INITIALIZER(0xCC,0xCC,0xCC, 0x00,0x00,0x55),
        .selchan   = NCCHANNELS_INITIALIZER(0x00,0x00,0x00, 0xFF,0xFF,0x00),
        .sepchan   = NCCHANNELS_INITIALIZER(0x88,0x88,0x88, 0x88,0x88,0x88),
        .separator = " | ",
        .flags     = NCTABBED_OPTION_BOTTOM,
    };                                                  // структура nctabbed_options :contentReference[oaicite:1]{index=1}

    struct nctabbed* tabs = nctabbed_create(host, &topts);
    // добавляем две вкладки
    struct nctab* t1 = nctabbed_add(tabs, NULL, NULL, tab1_cb, "Первая", NULL);
    struct nctab* t2 = nctabbed_add(tabs, t1, NULL, tab2_cb, "Вторая", NULL);

    // отрисовать и показать
    nctabbed_redraw(tabs);
    notcurses_render(nc);


    while (1) {}

    // чистим за собой
    nctabbed_destroy(tabs);
    notcurses_stop(nc);
    return EXIT_SUCCESS;
}


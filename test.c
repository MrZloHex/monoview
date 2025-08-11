#define _XOPEN_SOURCE 700
#include <locale.h>
#include <stdlib.h>
#include <stdio.h>
#include <wchar.h>
#include <notcurses/notcurses.h>

// пункты, которые добавим в селектор
static const char* labels[] = {
    "Option 1: One",
    "Option 2: Two",
    "Option 3: Three",
    "Option 4: Four",
    "Option 5: Five",
};
static const unsigned NLABELS = sizeof(labels)/sizeof(*labels);

int main(void){
    // требуется для корректного расчёта ширины UTF-8 символов :contentReference[oaicite:0]{index=0}
    setlocale(LC_ALL, "");

    struct notcurses* nc = notcurses_init(NULL, NULL);
    if(!nc){
        fprintf(stderr, "notcurses_init failed\n");
        return EXIT_FAILURE;
    }

    // создаём вспомогательную плоскость (смещаем на 1,1, остальное займёт селектор)
    struct ncplane_options popts = {
        .y = 1, .x = 1,
        .rows = 0, .cols = 0,
    };
    struct ncplane* p = ncplane_create(notcurses_stdplane(nc), &popts);
    if(!p){
        notcurses_stop(nc);
        return EXIT_FAILURE;
    }

    // создаём селектор без начальных items (будем добавлять сами),
    // flags = 0 → прокрутка НЕ зацикливается :contentReference[oaicite:1]{index=1}
    struct ncselector_options opts = {
        .title         = "Choose an option:",
        .secondary     = NULL,
        .footer        = "↑/↓ — листать, q — выход",
        .items         = NULL,
        .defidx        = 0,
        .maxdisplay    = 0,  // 0 → показывать все пункты
        .opchannels    = 0,
        .descchannels  = 0,
        .titlechannels = 0,
        .footchannels  = 0,
        .boxchannels   = 0,
        .flags         = 0,
    };
    struct ncselector* selector = ncselector_create(p, &opts);
    if(!selector){
        notcurses_stop(nc);
        return EXIT_FAILURE;
    }

    // динамически добавляем пункты в селектор :contentReference[oaicite:2]{index=2}
    for(unsigned i = 0; i < NLABELS; ++i){
        struct ncselector_item it = {
            .option = labels[i],
            .desc   = NULL,
        };
        if(ncselector_additem(selector, &it) < 0){
            fprintf(stderr, "failed to add item %u\n", i);
        }
    }

    // первый рендер
    notcurses_render(nc);

    // главный цикл: q → выход, стрелки и PgUp/PgDn — прокрутка без зацикливания
    struct ncinput ni;
    while(notcurses_get_blocking(nc, &ni) >= 0){
        if(ni.id == 'q'){
            break;
        }
        // если селектор обработал ввод (↑/↓/PgUp/PgDn/Enter), перерисовываем
        if(ncselector_offer_input(selector, &ni)){
            notcurses_render(nc);
            if(ni.id == '\n' || ni.id == NCKEY_ENTER){
                const char* sel = ncselector_selected(selector);
                fprintf(stderr, "You selected: %s\n", sel);
            }
        }
    }

    // уничтожаем селектор и завершаем
    ncselector_destroy(selector, NULL);
    notcurses_stop(nc);
    return EXIT_SUCCESS;
}


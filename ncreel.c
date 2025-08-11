#define _XOPEN_SOURCE 700
#include <locale.h>
#include <stdlib.h>
#include <stdio.h>
#include <notcurses/notcurses.h>

// callback, рисующий содержимое таблетки (только одну строку)
static int
draw_item(struct nctablet* t, bool focused){
    const char* txt = (const char*)nctablet_userptr(t);
    struct ncplane* p = nctablet_plane(t);
    ncplane_erase(p);
    ncplane_home(p);
    // делаем текст жирным для фокуса (опционально)
    ncplane_set_styles(p, focused ? NCSTYLE_BOLD : NCSTYLE_NONE);
    ncplane_putstr(p, txt);
    return 1; // возвращаем, сколько строк заняли
}

int
main(void){
    // важно для правильной работы wcwidth/wcswidth внутри Notcurses 
    setlocale(LC_ALL, "");

    struct notcurses* nc = notcurses_init(NULL, NULL);
    if(!nc){
        fprintf(stderr, "Не удалось инициализировать Notcurses\n");
        return EXIT_FAILURE;
    }
    struct ncplane* stdn = notcurses_stdplane(nc);

    // создаём ncreel с отключённой бесконечной прокруткой,
    // серым цветом границ для невыбранных и жёлтым для фокуса :contentReference[oaicite:1]{index=1}
    struct ncreel_options ropts = {
        .bordermask  = 0,
        .borderchan  = NCCHANNELS_INITIALIZER(100,100,100,  0,0,0),
        .tabletmask  = 0,
        .tabletchan  = NCCHANNELS_INITIALIZER(200,200,200,  0,0,0),
        .focusedchan = NCCHANNELS_INITIALIZER(255,255,0,    0,0,0),
        .flags       = 0
    };
    struct ncreel* reel = ncreel_create(stdn, &ropts);
    if(!reel){
        notcurses_stop(nc);
        return EXIT_FAILURE;
    }

    // добавляем 10 демонстрационных элементов
    for(int i = 1 ; i <= 10 ; ++i){
        char* buf = malloc(64);
        snprintf(buf, 64, "Item %2d: Lorem ipsum dolor sit amet", i);
        ncreel_add(reel, NULL, NULL, draw_item, buf);
    }

    // первый рендер
    ncreel_redraw(reel);
    notcurses_render(nc);

    // ввод: ↑/↓ — листаем, q/Esc — выход
    struct ncinput ni;
    while(notcurses_get_blocking(nc, &ni) >= 0){
        if(ni.id == NCKEY_UP || ni.id == NCKEY_DOWN){
            ncreel_offer_input(reel, &ni);
            notcurses_render(nc);
        }else if(ni.id == 'q' ){
            break;
        }
    }

    // очистка
    ncreel_destroy(reel);
    notcurses_stop(nc);
    return EXIT_SUCCESS;
}


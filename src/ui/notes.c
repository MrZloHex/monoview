#include "ui/notes.h"
#include "trace.h"

void
notes_init(Notes *nt, struct notcurses *nc)
{
    TRACE_INFO("UI init notes");
    
    unsigned y, x;
    struct ncplane* stdplane = notcurses_stddim_yx(nc, &y, &x);
    struct ncplane_options pots =
    {
        .y = 0,
        .x = 0,
        .rows = y-2,
        .cols = x,
    };
    nt->pl = ncplane_create(stdplane, &pots);
    ncplane_erase(nt->pl);

    cal_init(&nt->cal, nt->pl);

    pots = (struct ncplane_options)
    {
        .y = 10, .x = 0,
        .rows = y-2-10,
        .cols = 30,
    };
    nt->pl_filter = ncplane_create(nt->pl, &pots);
    uint64_t channels = NCCHANNELS_INITIALIZER(0xFF, 0xFF, 0xFF, 0x3A, 0x3A, 0xFF);
    ncplane_set_base(nt->pl_filter, " ", 0, channels);
    ncplane_erase(nt->pl_filter);

    pots = (struct ncplane_options)
    {
        .y = 0, .x = 30,
        .rows = y-2,
        .cols = x-30,
    };
    nt->pl_notes = ncplane_create(nt->pl, &pots);
    
    channels = NCCHANNELS_INITIALIZER(0xFF, 0xFF, 0xFF, 0x3A, 0xFA, 0x00);
    ncplane_set_base(nt->pl_notes, " ", 0, channels);
    ncplane_erase(nt->pl_notes);


    cal_init(&nt->cal, nt->pl);
    cal_render(&nt->cal, NULL);
}

void
notes_deinit(Notes *nt)
{
    TRACE_INFO("UI deinit notes");
    cal_deinit(&nt->cal);
    ncplane_destroy(nt->pl);
    ncplane_destroy(nt->pl_filter);
}

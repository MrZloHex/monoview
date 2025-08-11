#include "ui/notes.h"
#include "logic/notes.h"
#include "trace.h"

const char lorem[] = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.";

void
notes_init(Notes *nt, struct notcurses *nc)
{
    TRACE_INFO("UI init notes");
    nm_init(&nt->nm);
    nm_add_note(&nt->nm, lorem);
    nm_add_note(&nt->nm, lorem);
    nm_add_note(&nt->nm, lorem);
    
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
        .y = 11, .x = 0,
        .rows = y-2-11,
        .cols = 30,
    };
    nt->pl_filter = ncplane_create(nt->pl, &pots);
    uint64_t channels = NCCHANNELS_INITIALIZER(0xFF, 0xFF, 0xFF, 0x7c, 0x7f, 0x7d);
    ncplane_set_base(nt->pl_filter, " ", 0, channels);
    ncplane_erase(nt->pl_filter);

    pots = (struct ncplane_options)
    {
        .y = 0, .x = 30,
        .rows = y-2,
        .cols = x-30,
    };
    nt->pl_notes = ncplane_create(nt->pl, &pots);
    
    ncplane_erase(nt->pl_notes);
}

void
notes_render(Notes *nt)
{
    cal_render(&nt->cal, NULL);

    // NOTES RENDER
    ncplane_set_styles(nt->pl_notes, NCSTYLE_BOLD);
    ncplane_printf_yx(nt->pl_notes, 2, 10, "Maybe");
    ncplane_set_styles(nt->pl_notes, 0);
    size_t b;
    ncplane_puttext(nt->pl_notes, 4, NCALIGN_CENTER, "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.", &b);
    // NOTES RENDER
}

void
notes_input(Notes *nt, ncinput in)
{
    (void)in;
}

void
notes_deinit(Notes *nt)
{
    TRACE_INFO("UI deinit notes");
    cal_deinit(&nt->cal);
    ncplane_destroy(nt->pl);
    ncplane_destroy(nt->pl_filter);
}

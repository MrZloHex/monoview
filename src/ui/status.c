#include "ui/status.h"
#include "trace.h"

void
sb_init(StatusBar *sb, struct notcurses *nc)
{
    TRACE_INFO("UI init status bar");
    unsigned y, x;
    struct ncplane* stdplane = notcurses_stddim_yx(nc, &y, &x);
    struct ncplane_options pots =
    {
        .y = y - 2,
        .x = 0,
        .rows = 1,
        .cols = x,
    };
    sb->pl = ncplane_create(stdplane, &pots);
    uint64_t channels = NCCHANNELS_INITIALIZER(0xFF, 0xFF, 0xFF, 0x3A, 0x3A, 0x3A);
    ncplane_set_base(sb->pl, " ", 0, channels);
    ncplane_erase(sb->pl);
    ncplane_set_fg_rgb(sb->pl, 0xfed6ae);
    ncplane_set_bg_rgb(sb->pl, 0x949494);
        
}

void
sb_deinit(StatusBar *sb)
{
    TRACE_INFO("UI deinit status bar");
    ncplane_destroy(sb->pl);
}

#if 0
void
sb_enable_cmd(StatusBar *sb)
{
    struct ncplane_options pots =
    {
        .rows = 1,
        .cols = 5,
    };
    sb->cmd = ncplane_create(sb->pl, &pots);
    ncplane_erase(sb->cmd);
    ncplane_set_fg_rgb(sb->cmd, 0xfed6ae);
    ncplane_set_bg_rgb(sb->cmd, 0x87af87);
    ncplane_putstr_yx(sb->cmd, 0, 0, " CMD ");
}
void
sb_disable_cmd(StatusBar *sb)
{
    ncplane_destroy(sb->cmd);
    ncplane_erase(sb->pl);
}
#endif

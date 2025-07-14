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
        
    
    pots = (struct ncplane_options)
    {
        .y = 0,
        .x = 0,
        .rows = 1,
        .cols = 5,
    };
    sb->status = ncplane_create(sb->pl, &pots);
    uint64_t ch = NCCHANNELS_INITIALIZER(0xfb, 0xd6, 0xae, 0x87, 0xAF, 0x87);
    ncchannels_set_bg_alpha(&ch, NCALPHA_TRANSPARENT );
    ncplane_set_base(sb->status, " ", 0, ch);
    ncplane_erase(sb->status);
}

void
sb_deinit(StatusBar *sb)
{
    TRACE_INFO("UI deinit status bar");
    ncplane_destroy(sb->pl);
    ncplane_destroy(sb->status);
}

void
sb_enable_cmd(StatusBar *sb)
{
    TRACE_INFO("UI SB enable cmd");
    sb->in_cmd_mode = true;
    uint64_t ch = NCCHANNELS_INITIALIZER(0xfb, 0xd6, 0xae, 0x87, 0xAF, 0x87);
    ncplane_set_base(sb->status, " ", 0, ch);
    ncplane_putstr_yx(sb->status, 0, 0, " CMD ");
}
void
sb_disable_cmd(StatusBar *sb)
{
    TRACE_INFO("UI SB disable cmd");
    sb->in_cmd_mode = false;
    uint64_t ch = NCCHANNELS_INITIALIZER(0xfb, 0xd6, 0xae, 0x87, 0xAF, 0x87);
    ncchannels_set_bg_alpha(&ch, NCALPHA_TRANSPARENT );
    ncplane_set_base(sb->status, " ", 0, ch);
    ncplane_erase(sb->status);

}

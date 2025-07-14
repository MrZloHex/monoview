#include "ui/ui.h"
#include "ui/util.h"
#include "trace.h"

void *
ui_thread(void *arg)
{
    (void)arg;

    UI ui = { 0 };

    notcurses_options ncopt;
    memset(&ncopt, 0, sizeof(ncopt));
    ui.nc = notcurses_init(&ncopt, stdout);
    if (!ui.nc)
    { TRACE_FATAL("Failed to init NOTCURSES"); }

    notcurses_linesigs_disable(ui.nc);
    notcurses_mice_disable(ui.nc);

    sb_init(&ui.sb, ui.nc);
    cmd_init(&ui.cmd, ui.nc);

    notcurses_render(ui.nc);

    ncinput in;
    while (!ui.should_close)
    {
        notcurses_get_blocking(ui.nc, &in);
        ncinput_dump(in);
        
        if (ui.mod == MOD_REG)
        {
            if (in.id == 'q')
            { ui.should_close = true; }
        }
        
        notcurses_render(ui.nc);
    }

    sb_deinit(&ui.sb);
    cmd_deinit(&ui.cmd);

    notcurses_stop(ui.nc);

    TRACE_INFO("FINISHING WITH NOTCURSES");
    return NULL;
}


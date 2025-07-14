#include "ui/cmd.h"
#include "trace.h"

void
cmd_init(CMD *cmd, struct notcurses *nc)
{
    TRACE_INFO("UI init cmd");
    
    unsigned y, x;
    struct ncplane* stdplane = notcurses_stddim_yx(nc, &y, &x);
    struct ncplane_options pots =
    {
        .y = y - 1,
        .x = 0,
        .rows = 1,
        .cols = x,
    };
    cmd->pl = ncplane_create(stdplane, &pots);
    ncplane_erase(cmd->pl);

    struct ncreader_options rots =
    {
        .flags = NCREADER_OPTION_HORSCROLL | NCREADER_OPTION_CURSOR
    };
    cmd->rd = ncreader_create(cmd->pl, &rots);
}

void
cmd_deinit(CMD *cmd)
{
    TRACE_INFO("UI deinit cmd");
    ncplane_destroy(cmd->pl);
    ncreader_destroy(cmd->rd, NULL);
}

#if 0

void
cmd_input(UI *ui, ncinput ni)
{
    if (ni.evtype != NCTYPE_PRESS)
    { return; }

    CMD *cmd = &ui->cmd;

    if (ni.id == NCKEY_ENTER)
    {
        const char* c = ncreader_contents(cmd->rd);
        TRACE_INFO("CMD `%s`", c);
        if (strcmp(c, "q") == 0)
        { ui->should_close = true; }

        cmd_deinit(ui);
    }
    else
    {
        ncreader_offer_input(cmd->rd, &ni);
    }


}
#endif

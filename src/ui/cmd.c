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

    pots = (struct ncplane_options)
    {
        .y = 0,
        .x = 1,
        .rows = 1,
        .cols = x-1,
    };
    cmd->rd_pl = ncplane_create(cmd->pl, &pots);
    ncplane_erase(cmd->rd_pl);

    struct ncreader_options rots =
    {
        .flags = NCREADER_OPTION_HORSCROLL | NCREADER_OPTION_CURSOR
    };
    cmd->rd = ncreader_create(cmd->rd_pl, &rots);
}

void
cmd_deinit(CMD *cmd)
{
    TRACE_INFO("UI deinit cmd");
    ncplane_destroy(cmd->pl);
    ncplane_destroy(cmd->rd_pl);
    ncreader_destroy(cmd->rd, NULL);
}

void
cmd_enable(CMD *cmd)
{
    ncplane_putchar_yx(cmd->pl, 0, 0, ':');
}

void
cmd_disable(CMD *cmd)
{
    ncreader_clear(cmd->rd);
    ncplane_erase(cmd->pl);
}


Command
cmd_input(CMD *cmd, ncinput ni)
{
    if (ni.id == NCKEY_ENTER && ni.evtype == NCTYPE_PRESS)
    {
        const char* c = ncreader_contents(cmd->rd);
        TRACE_INFO("CMD `%s`", c);
        
        if (strcmp(c, "q") == 0)
        { return (Command) { .kind = CMD_QUIT }; }

        return (Command) { .kind = CMD_UNKNOWN };
    }

    ncreader_offer_input(cmd->rd, &ni);
    return (Command) { .kind = CMD_NOP };
}

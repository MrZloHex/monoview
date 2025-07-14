#ifndef __UI_CMD_H__
#define __UI_CMD_H__

#include <notcurses/notcurses.h>

typedef struct
{
    struct ncplane  *pl;
    struct ncplane  *rd_pl;
    struct ncreader *rd;
} CMD;

typedef struct
{
    enum
    {
        CMD_NOP,
        CMD_UNKNOWN,

        CMD_QUIT
        
    } kind;
} Command;

void
cmd_init(CMD *cmd, struct notcurses *nc);

void
cmd_deinit(CMD *cmd);

void
cmd_enable(CMD *cmd);

void
cmd_disable(CMD *cmd);

Command
cmd_input(CMD *cmd, ncinput ni);

#endif /* __UI_CMD_H__ */

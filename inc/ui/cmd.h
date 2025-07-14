#ifndef __UI_CMD_H__
#define __UI_CMD_H__

#include <notcurses/notcurses.h>

typedef struct
{
    struct ncplane  *pl;
    struct ncreader *rd;
} CMD;

void
cmd_init(CMD *cmd, struct notcurses *nc);

void
cmd_deinit(CMD *cmd);

void
cmd_input(CMD *cmd, ncinput ni);

#endif /* __UI_CMD_H__ */

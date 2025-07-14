#ifndef __UI_NOTES_H__
#define __UI_NOTES_H__

#include <notcurses/notcurses.h>
#include "ui/calendar.h"

typedef struct
{
    struct ncplane *pl;
    struct ncplane *pl_filter;
    struct ncplane *pl_notes;

    Calendar        cal;
} Notes;

void
notes_init(Notes *nt, struct notcurses *nc);

void
notes_deinit(Notes *nt);

#endif /* __UI_NOTES_H__ */

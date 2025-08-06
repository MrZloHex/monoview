#include "logic/logic.h"

#include "trace.h"

const char lorem[] = "LOREM IMPSUM";

void *
logic_thread(void *arg)
{
    (void)arg;

    Logic logic = { 0 };

    nm_init(&logic.notes);
    nm_add_note(&logic.notes, lorem);
    nm_add_note(&logic.notes, lorem);
    nm_add_note(&logic.notes, lorem);

    while (1)
    {
    }

    nm_deinit(&logic.notes);

    return NULL;
}


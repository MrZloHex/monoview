#include "logic/notes.h"

#include <stdlib.h>
#include <string.h>
#include "trace.h"

void
nm_init(NoteManager *nm)
{
    TRACE_INFO("LOGIC NM init");
    nm->master_id = 1;
    na_init(&nm->notes, 16);
}

void
nm_add_note(NoteManager *nm, const char *text)
{
    Note n =
    {
        .text = malloc(strlen(text)+1),
        .id   = nm->master_id++,
        .timestamp = time(NULL)
    };
    if (!n.text)
    { TRACE_FATAL("Out of memory"); }
    strcpy(n.text, text);
    
    na_append(&nm->notes, n);
}

void
nm_get_notes(NoteManager *nm, NoteArray *arr)
{
}

void
nm_deinit(NoteManager *nm)
{
    TRACE_INFO("LOGIC NM deinit");
    for (size_t i = 0; i < nm->notes.size; ++i)
    { free(nm->notes.data[i].text); }
    na_deinit(&nm->notes);
}

DEFINE_DYNARRAY(na, NoteArray, Note);

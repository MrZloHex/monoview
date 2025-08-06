#ifndef __LOGIC_NOTES_H__
#define __LOGIC_NOTES_H__

#include <pthread.h>
#include <time.h>
#include "dynarray.h"

typedef struct
{
    int id;
    time_t timestamp;
    char *text;
} Note;

DECLARE_DYNARRAY(na, NoteArray, Note);

typedef struct
{
    NoteArray        notes;
    int              master_id;
} NoteManager;

void
nm_init(NoteManager *nm);

void
nm_add_note(NoteManager *nm, const char *text);

void
nm_deinit(NoteManager *nm);

#endif /* __LOGIC_NOTES_H__ */

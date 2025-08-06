#ifndef __LOGIC_H__
#define __LOGIC_H__

#include "logic/notes.h"

typedef struct Logic
{
    NoteManager notes;
} Logic;

void *
logic_thread(void *q);

#endif /* __LOGIC_H__ */

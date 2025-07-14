#ifndef __UI_CALENDAR_H__
#define __UI_CALENDAR_H__

#include <notcurses/notcurses.h>

typedef struct
{
    struct ncplane *pl;
} Calendar;

typedef struct
{
    uint8_t date;
    uint8_t month;
    uint16_t year;
} Date;

void
cal_init(Calendar *cal, struct ncplane *pl);

void
cal_deinit(Calendar *cal);

void
cal_render(Calendar *cal, Date *high);

#endif /* __UI_CALENDAR_H__ */

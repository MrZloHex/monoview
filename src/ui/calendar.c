#include "ui/calendar.h"

#include <time.h>
#include "trace.h"

void
cal_init(Calendar *cal, struct ncplane *pl)
{
    unsigned y, x; ncplane_dim_yx(pl, &y, &x);
    struct ncplane_options pots =
    {
        .y = 0,
        .x = 0,
        .rows = 11,
        .cols = 30,
    };
    cal->pl = ncplane_create(pl, &pots);
    uint64_t channels = NCCHANNELS_INITIALIZER(0xFF, 0xFF, 0xFF, 0x1c, 0x7f, 0x7d);
    ncplane_set_base(cal->pl, " ", 0, channels);
    ncplane_erase(cal->pl);
}

static const char *month_str[] =
{ "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec" };

static const uint8_t month_days[] =
{ 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31 };

static inline Date
prev_month(Date d)
{
    return (Date)
    {
        .year  = (d.month == 0) ? d.year-1 : d.year,
        .month = (d.month == 0) ? 11 : d.month -1,
        .date  = d.date
    };
}

static uint8_t
days_month(Date date)
{
    if (date.month == 1)
    {
        return ((date.year % 4 == 0 && date.year % 100 != 0)
                || (date.year % 400 == 0)) ? 29 : 28;
    }
    return month_days[date.month];
}

static uint8_t
wday_month(Date date)
{
    struct tm time_in = {0};
    time_in.tm_year = date.year - 1900;
    time_in.tm_mon  = date.month;
    time_in.tm_mday = 1;
    mktime(&time_in);
    return (time_in.tm_wday + 6) % 7;
}

void
cal_render(Calendar *cal, Date *high)
{
    time_t now = time(NULL);
    struct tm *t = localtime(&now);

    Date date = { 0 };
    if (!high)
    {
        date = (Date)
        {
            .year = t->tm_year + 1900,
            .month = t->tm_mon,
            .date = t->tm_mday,
        };
    }
    else
    { date = *high; }

    ncplane_printf_yx(cal->pl, 1, 10, "%s %u", month_str[date.month], date.year);

    ncplane_set_styles(cal->pl, NCSTYLE_ITALIC);
    ncplane_set_fg_rgb8(cal->pl, 0x7c, 0x7f, 0x7e);
    ncplane_putstr_yx(cal->pl, 3, 4, "Mo Tu We Th Fr Sa Su");
    ncplane_set_fg_default(cal->pl);
    ncplane_set_styles(cal->pl, 0);

    ncplane_cursor_move_yx(cal->pl, 4, 4);

    uint8_t wday_1st = wday_month(date);
    if (wday_1st != 0)
    {
        uint8_t days_prev = days_month(prev_month(date));
        TRACE_INFO("PREV %u %u", days_prev, wday_1st);
        ncplane_set_fg_rgb8(cal->pl, 0x5e, 0x57, 0x53);
        for (uint8_t d = days_prev+1 - wday_1st; d <= days_prev; ++d)
        { ncplane_printf(cal->pl, "%u ", d); }
        ncplane_set_fg_default(cal->pl);
    }
    
    bool high_today = (date.month == t->tm_mon && date.year == t->tm_year+1900);
    int row = 4;
    uint8_t wd = wday_1st;

    for (uint8_t d = 1; d <= days_month(date); ++d, ++wd)
    {
        if (wd == 7)
        {
            ncplane_cursor_move_yx(cal->pl, ++row, 4);
            wd = 0;
        }

        if (high_today && d == t->tm_mday)
        { ncplane_set_styles(cal->pl, NCSTYLE_UNDERLINE); }
        if (d == date.date)
        { ncplane_set_bg_rgb8(cal->pl, 0x72, 0x5d, 0x2a); }
        if (wd == 5 || wd == 6)
        { ncplane_set_fg_rgb8(cal->pl, 0x86, 0x40, 0x15); }

        ncplane_printf(cal->pl, "%02u", d);

        if (high_today && d == t->tm_mday)
        { ncplane_set_styles(cal->pl, 0); }
        if (d == date.date)
        { ncplane_set_bg_default(cal->pl); }
        if (wd == 5 || wd == 6)
        { ncplane_set_fg_default(cal->pl); }

        ncplane_putchar(cal->pl, ' ');
    }

    if (wd != 7)
    {
        ncplane_set_fg_rgb8(cal->pl, 0x5e, 0x57, 0x53);
        for (uint8_t d = 1; wd < 7; ++d, ++wd)
        { ncplane_printf(cal->pl, "%02u ", d); }
        ncplane_set_fg_default(cal->pl);
    }
}

void
cal_deinit(Calendar *cal)
{
    ncplane_destroy(cal->pl);
}

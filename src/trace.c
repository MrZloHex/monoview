/* =============================================================================
 *                          ░▒▓█ _TMA_ █▓▒░
 *
 *   File       : trace.c
 *   Author     : MrZloHex
 *   Date       : 2025-02-01
 *
 *   Description: 
 *      Implementation of logging and profiling functions.
 *
 *  This file implements:
 *      - Logger initialization and configuration.
 *      - Formatted log output to the console.
 *
 * =============================================================================
 */

#include "trace.h"

#include <stdio.h>
#include <stdarg.h>
#include <stdint.h>
#include <string.h>
#include <time.h>


Tracer tracer;

static const char *const trace_levels[] =
{
    "DEBUG", "INFO", "WARN", "ERROR", "FATAL"
};

static const char *const trace_colors[] =
{
    "\x1b[34m",  // DEBUG: Blue
    "\x1b[32m",  // INFO:  Green
    "\x1b[33m",  // WARN:  Yellow
    "\x1b[31m",  // ERROR: Red
    "\x1b[35m"   // FATAL: Magenta
};

static int
get_datetime(char *buffer, size_t size)
{
    time_t t = time(NULL);
    struct tm tm_info;
#if defined(_POSIX_C_SOURCE) && _POSIX_C_SOURCE >= 1
    localtime_r(&t, &tm_info);
#else
    struct tm *tmp = localtime(&t);
    if (tmp)
    { tm_info = *tmp; }
#endif

    return strftime(buffer, size, "%Y-%m-%d %H:%M:%S", &tm_info);
}

void
tracer_init(Trace_Level level, Trace_Params params)
{
    tracer.base_level = level;
    tracer.params     = params;
}

void
tracer_set_level(Trace_Level level)
{
    tracer.base_level = level;
}

static void
_tracer_print(Trace_Entry entry)
{
    printf("%s\n", entry.fmt);
    fflush(stdout);
}

void
tracer_trace
(
    Trace_Level level, const char *restrict file,
    const char *restrict func, int line, const char *fmt, ...
)
{
    if (tracer.base_level > level)
    { return; }

    va_list args;
    va_start(args, fmt);

    Trace_Entry entry = { .lvl = level };
    int pc = 0;

    if (IS_ENABLED(TP_TIME))
    {
        char datetime[64];
        get_datetime(datetime, sizeof(datetime));
        pc += snprintf(entry.fmt, MAX_ENTRY_DSCR_LEN, "[%s]", datetime);
    }
    pc += snprintf(entry.fmt + pc, MAX_ENTRY_DSCR_LEN - pc, " %s%5s\x1b[0m ", trace_colors[level], trace_levels[level]);

    if (IS_ENABLED(TP_FILE))
        pc += snprintf(entry.fmt + pc, MAX_ENTRY_DSCR_LEN - pc, "%s:", file);
    if (IS_ENABLED(TP_FUNC))
        pc += snprintf(entry.fmt + pc, MAX_ENTRY_DSCR_LEN - pc, "%s:", func);
    if (IS_ENABLED(TP_LINE))
        pc += snprintf(entry.fmt + pc, MAX_ENTRY_DSCR_LEN - pc, "%d: ", line);

    pc += vsnprintf(entry.fmt + pc, MAX_ENTRY_DSCR_LEN - pc, fmt, args);
    va_end(args);

    _tracer_print(entry);
}


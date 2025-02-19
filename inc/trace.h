/* =============================================================================
 *                          ░▒▓█ _TMA_ █▓▒░
 *
 *   File       : trace.h
 *   Author     : MrZloHex
 *   Date       : 2025-02-01
 *
 *   Description:
 *      A library for logging and profiling on STM32 microcontrollers.
 *
 *      This library provides a universal mechanism for outputting debug information,
 *      measuring code execution time using DWT (Data Watchpoint and Trace),
 *      as well as optional caching of logs in MRAM memory (if available) for subsequent analysis.
 *
 *   This library provides a mechanism for outputting debug information,
 *   measuring code execution time using the CPU cycle counter (via RDTSC),
 *   and formatting log messages with contextual information.
 *
 * =============================================================================
 */

#ifndef __TRACE_H__
#define __TRACE_H__

#include <stdint.h>
#include <stdarg.h>
#include <sys/time.h>

typedef enum
{
    TRC_DEBUG,
    TRC_INFO,
    TRC_WARN,
    TRC_ERROR,
    TRC_FATAL
} Trace_Level;

/** The maximum length of the formatted log string. */
#define MAX_ENTRY_DSCR_LEN 512

typedef struct
{
    Trace_Level lvl;
    char fmt[MAX_ENTRY_DSCR_LEN];
} Trace_Entry;

typedef enum
{
    TP_FILE = 0b0001,
    TP_FUNC = 0b0010,
    TP_LINE = 0b0100,
    TP_TIME = 0b1000,
    TP_ALL  = 0b1111
} Trace_Params;

/** Macro for checking output parameters. */
#define IS_ENABLED(__TP__) (tracer.params & (__TP__))

typedef struct Tracer_S
{
    Trace_Level  base_level;
    Trace_Params params;
} Tracer;

extern Tracer tracer;

/* Convenience macro to extract the filename (strips the path) */
#define __FILENAME__ (__builtin_strrchr("/" __FILE__, '/') + 1)

/**
 * @brief Logger initialization.
 *
 * @param level  Base logging level.
 * @param params Output parameters (e.g., file name, function, line).
 */
void
tracer_init(Trace_Level level, Trace_Params params);

/**
 * @brief Set the base logging level.
 *
 * @param level The new logging level.
 */
void
tracer_set_level(Trace_Level level);

/**
 * @brief Output a log message with a given level and format.
 *
 * @param level The log message level.
 * @param file  The name of the file where the message is formed.
 * @param func  The name of the function.
 * @param line  The line number in the file.
 * @param fmt   Formatted string and associated parameters.
 */
void
tracer_trace
(
    Trace_Level level, const char *restrict file,
    const char *restrict func, int line, const char *fmt, ...
);

/**
 * Macros for quick logging of messages at various levels.
 * These automatically substitute the file name, function, and line number.
 */
#define TRACE_DEBUG(...) tracer_trace(TRC_DEBUG, __FILENAME__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_INFO(...)  tracer_trace(TRC_INFO,  __FILENAME__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_WARN(...)  tracer_trace(TRC_WARN,  __FILENAME__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_ERROR(...) tracer_trace(TRC_ERROR, __FILENAME__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_FATAL(...) tracer_trace(TRC_FATAL, __FILENAME__, __func__, __LINE__, __VA_ARGS__)

#endif /* __TRACE_H__ */


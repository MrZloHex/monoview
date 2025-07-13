/* =============================================================================
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
#include <stdio.h>
#include <pthread.h>
#include <stdatomic.h>

#define TRACER_MODE_BLOCKING 0
#define TRACER_MODE_ASYNC    1

#define TRACER_MODE TRACER_MODE_ASYNC

#ifndef TRACER_MODE
#  error "TRACER_MODE must be defined"
#endif
#if TRACER_MODE != TRACER_MODE_BLOCKING && TRACER_MODE != TRACER_MODE_ASYNC
#  error "Invalid TRACER_MODE"
#endif

typedef enum
{
    TRC_DEBUG,
    TRC_INFO,
    TRC_WARN,
    TRC_ERROR,
    TRC_FATAL
} Trace_Level;

typedef enum
{
    TO_STDOUT = 0x1,
    TO_STDERR = 0x2
} Trace_Output;

typedef enum
{
    TP_FILE = 0x1,
    TP_FUNC = 0x2,
    TP_LINE = 0x4,
    TP_TIME = 0x8,
    TP_ALL  = TP_FILE | TP_FUNC | TP_LINE | TP_TIME
} Trace_Params;

#define IS_ENABLED(p)   ((tracer.params & (p)) != 0)
#define MAX_ENTRY_LEN   512

typedef struct LogItem_S
{
    struct LogItem_S *next;
    char              msg[MAX_ENTRY_LEN];
} LogItem;

typedef struct Tracer_S
{
    Trace_Level        base_level;
    Trace_Params       params;
    FILE             **streams;
    size_t             stream_count;
    pthread_rwlock_t   streams_lock;
#if TRACER_MODE == TRACER_MODE_BLOCKING
    pthread_mutex_t    lock;
#else
    _Atomic(LogItem *) head;
    pthread_t          worker;
    atomic_bool        running;
    pthread_mutex_t    cond_mutex;
    pthread_cond_t     cond;
#endif
} Tracer;

extern Tracer tracer;

int
tracer_init
(
    Trace_Level  level,
    Trace_Params params,
    int          outputs
);

void
tracer_set_level(Trace_Level level);

int
tracer_add_stream(FILE *stream);

int
tracer_remove_stream(FILE *stream);

#if TRACER_MODE == TRACER_MODE_ASYNC
int
tracer_start(void);

void
tracer_stop(void);
#endif

void
tracer_trace
(
    Trace_Level  level,
    const char  *file,
    const char  *func,
    int          line,
    const char  *fmt,
    ...
);

#define TRACE_DEBUG(...) tracer_trace(TRC_DEBUG, __FILE__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_INFO(...)  tracer_trace(TRC_INFO,  __FILE__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_WARN(...)  tracer_trace(TRC_WARN,  __FILE__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_ERROR(...) tracer_trace(TRC_ERROR, __FILE__, __func__, __LINE__, __VA_ARGS__)
#define TRACE_FATAL(...) tracer_trace(TRC_FATAL, __FILE__, __func__, __LINE__, __VA_ARGS__)

#endif /* __TRACE_H__ */


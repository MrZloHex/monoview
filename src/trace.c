#include "trace.h"
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/time.h>

Tracer tracer;

static const char *const trace_levels[] =
{
    "DEBUG",
    "INFO",
    "WARN",
    "ERROR",
    "FATAL"
};

static const char *const trace_colors[] =
{
    "\x1b[34m",
    "\x1b[32m",
    "\x1b[33m",
    "\x1b[31m",
    "\x1b[35m"
};

static int
get_datetime
(
    char   *buf,
    size_t  sz
)
{
    struct timeval tv;
    struct tm      tm_info;
    int            len;

    gettimeofday(&tv, NULL);
    localtime_r(&tv.tv_sec, &tm_info);
    len = strftime(buf, sz, "%Y-%m-%d %H:%M:%S", &tm_info);
    len += snprintf(buf + len, sz - len, ".%03ld", tv.tv_usec / 1000);
    return len;
}

int
tracer_init
(
    Trace_Level  level,
    Trace_Params params,
    int          outputs
)
{
    if (pthread_rwlock_init(&tracer.streams_lock, NULL) != 0)
    { return -1; }

#if TRACER_MODE == TRACER_MODE_BLOCKING
    if (pthread_mutex_init(&tracer.lock, NULL) != 0)
    { return -1; }
#else
    atomic_store_explicit(&tracer.head, NULL, memory_order_relaxed);
    atomic_store_explicit(&tracer.running, true, memory_order_release);
    if (pthread_mutex_init(&tracer.cond_mutex, NULL) != 0)
    { return -1; }
    if (pthread_cond_init(&tracer.cond, NULL) != 0)
    { return -1; }
    if (tracer_start() != 0)
    { return -1; }
#endif

    tracer.base_level   = level;
    tracer.params       = params;
    tracer.stream_count = 0;
    tracer.streams      = NULL;

    if (outputs & TO_STDOUT)
    { tracer_add_stream(stdout); }
    if (outputs & TO_STDERR)
    { tracer_add_stream(stderr); }

    return 0;
}

void
tracer_set_level(Trace_Level level)
{ tracer.base_level = level; }

int
tracer_add_stream(FILE *stream)
{
    pthread_rwlock_wrlock(&tracer.streams_lock);
    for (size_t i = 0; i < tracer.stream_count; ++i)
    {
        if (tracer.streams[i] == stream)
        {
            pthread_rwlock_unlock(&tracer.streams_lock);
            return 0;
        }
    }

    FILE **arr = realloc(tracer.streams,
                         (tracer.stream_count + 1) * sizeof(FILE*));
    if (!arr)
    {
        pthread_rwlock_unlock(&tracer.streams_lock);
        return -1;
    }

    tracer.streams                        = arr;
    tracer.streams[tracer.stream_count++] = stream;
    pthread_rwlock_unlock(&tracer.streams_lock);

    return 0;
}

int
tracer_remove_stream(FILE *stream)
{
    pthread_rwlock_wrlock(&tracer.streams_lock);

    size_t idx = tracer.stream_count;
    for (size_t i = 0; i < tracer.stream_count; ++i)
    {
        if (tracer.streams[i] == stream)
        {
            idx = i;
            break;
        }
    }
    if (idx == tracer.stream_count)
    {
        pthread_rwlock_unlock(&tracer.streams_lock);
        return -1;
    }

    memmove(&tracer.streams[idx],
            &tracer.streams[idx + 1],
            (tracer.stream_count - idx - 1) * sizeof(FILE*));
    tracer.stream_count -= 1;
    pthread_rwlock_unlock(&tracer.streams_lock);

    return 0;
}

#if TRACER_MODE == TRACER_MODE_ASYNC
static void*
logger_thread(void *arg)
{
    (void)arg;
    while (atomic_load_explicit(&tracer.running, memory_order_acquire))
    {
        LogItem *list = atomic_exchange_explicit(&tracer.head,
                                                 (LogItem*)NULL,
                                                 memory_order_acq_rel);
        if (!list)
        {
            pthread_mutex_lock(&tracer.cond_mutex);
            pthread_cond_wait(&tracer.cond, &tracer.cond_mutex);
            pthread_mutex_unlock(&tracer.cond_mutex);
            continue;
        }
        while (list)
        {
            LogItem *next = list->next;
            pthread_rwlock_rdlock(&tracer.streams_lock);
            for (size_t i = 0; i < tracer.stream_count; ++i)
            {
                fputs(list->msg, tracer.streams[i]);
                fputc('\n',   tracer.streams[i]);
                fflush(tracer.streams[i]);
            }
            pthread_rwlock_unlock(&tracer.streams_lock);
            free(list);
            list = next;
        }
    }
    return NULL;
}

int
tracer_start(void)
{
    if (pthread_create(&tracer.worker,
                       NULL,
                       logger_thread,
                       NULL) != 0)
    {
        atomic_store_explicit(&tracer.running,
                              false,
                              memory_order_release);
        return -1;
    }
    return 0;
}

void
tracer_stop(void)
{
    atomic_store_explicit(&tracer.running,
                          false,
                          memory_order_release);

    pthread_mutex_lock(&tracer.cond_mutex);
    pthread_cond_signal(&tracer.cond);
    pthread_mutex_unlock(&tracer.cond_mutex);

    pthread_join(tracer.worker, NULL);
    pthread_cond_destroy(&tracer.cond);
    pthread_mutex_destroy(&tracer.cond_mutex);

    LogItem *list = atomic_load_explicit(&tracer.head,
                                         memory_order_acquire);
    while (list)
    {
        LogItem *next = list->next;
        free(list);
        list = next;
    }
}
#endif /* TRACER_MODE_ASYNC */

void
tracer_trace
(
    Trace_Level  level,
    const char  *file,
    const char  *func,
    int          line,
    const char  *fmt,
    ...
)
{
    if (level < tracer.base_level)
        return;

    char    buf[MAX_ENTRY_LEN];
    int     pos = 0;
    va_list args;

    va_start(args, fmt);
    if (IS_ENABLED(TP_TIME))
    {
        pos += get_datetime(buf, sizeof(buf));
        buf[pos++] = ' ';
    }
    pos += snprintf(buf + pos,
                    sizeof(buf) - pos,
                    "%s%5s\x1b[0m ",
                    trace_colors[level],
                    trace_levels[level]);

    if (IS_ENABLED(TP_FILE))
        pos += snprintf(buf + pos,
                        sizeof(buf) - pos,
                        "%s:",
                        file);
    if (IS_ENABLED(TP_FUNC))
        pos += snprintf(buf + pos,
                        sizeof(buf) - pos,
                        "%s:",
                        func);
    if (IS_ENABLED(TP_LINE))
        pos += snprintf(buf + pos,
                        sizeof(buf) - pos,
                        "%d: ",
                        line);

    pos += vsnprintf(buf + pos,
                     sizeof(buf) - pos,
                     fmt,
                     args);

#if TRACER_MODE == TRACER_MODE_ASYNC
    LogItem *node = malloc(sizeof(LogItem));
    if (!node)
    { return; }

    size_t len = strnlen(buf, sizeof(node->msg) - 1);
    memcpy(node->msg, buf, len);
    node->msg[len] = '\0';

    LogItem *old_head;
    do
    {
        old_head = atomic_load_explicit(&tracer.head,
                                        memory_order_relaxed);
        node->next = old_head;
    } while (!atomic_compare_exchange_weak_explicit(
             &tracer.head,
             &old_head,
             node,
             memory_order_acq_rel,
             memory_order_relaxed));

    pthread_mutex_lock(&tracer.cond_mutex);
    pthread_cond_signal(&tracer.cond);
    pthread_mutex_unlock(&tracer.cond_mutex);
#else
    pthread_mutex_lock(&tracer.lock);
    pthread_rwlock_rdlock(&tracer.streams_lock);
    for (size_t i = 0; i < tracer.stream_count; ++i)
    {
        fputs(buf, tracer.streams[i]);
        fputc('\n', tracer.streams[i]);
        fflush(tracer.streams[i]);
    }
    pthread_rwlock_unlock(&tracer.streams_lock);
    pthread_mutex_unlock(&tracer.lock);
#endif
}


#include "queue.h"
#include <stdlib.h>

int queue_init(queue_t *q, size_t cap) {
    q->buffer = calloc(cap, sizeof(message_t));
    if (!q->buffer) return -1;
    q->cap = cap;
    q->head = q->tail = 0;
    pthread_mutex_init(&q->mtx, NULL);
    pthread_cond_init(&q->not_empty, NULL);
    pthread_cond_init(&q->not_full, NULL);
    return 0;
}

void queue_destroy(queue_t *q) {
    free(q->buffer);
    pthread_mutex_destroy(&q->mtx);
    pthread_cond_destroy(&q->not_empty);
    pthread_cond_destroy(&q->not_full);
}

void queue_push(queue_t *q, message_t msg) {
    pthread_mutex_lock(&q->mtx);
    while ((q->tail + 1) % q->cap == q->head)
        pthread_cond_wait(&q->not_full, &q->mtx);
    q->buffer[q->tail] = msg;
    q->tail = (q->tail + 1) % q->cap;
    pthread_cond_signal(&q->not_empty);
    pthread_mutex_unlock(&q->mtx);
}

int queue_try_pop(queue_t *q, message_t *msg) {
    int ret = -1;
    pthread_mutex_lock(&q->mtx);
    if (q->head != q->tail) {
        *msg = q->buffer[q->head];
        q->head = (q->head + 1) % q->cap;
        pthread_cond_signal(&q->not_full);
        ret = 0;
    }
    pthread_mutex_unlock(&q->mtx);
    return ret;
}

message_t queue_pop(queue_t *q) {
    message_t msg;
    pthread_mutex_lock(&q->mtx);
    while (q->head == q->tail)
        pthread_cond_wait(&q->not_empty, &q->mtx);
    msg = q->buffer[q->head];
    q->head = (q->head + 1) % q->cap;
    pthread_cond_signal(&q->not_full);
    pthread_mutex_unlock(&q->mtx);
    return msg;
}


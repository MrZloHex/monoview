#ifndef QUEUE_H
#define QUEUE_H
#include <pthread.h>
#include <stdlib.h>

typedef struct {
    int type;
    void *payload;
} message_t;

typedef struct {
    message_t *buffer;
    size_t head, tail, cap;
    pthread_mutex_t mtx;
    pthread_cond_t not_empty, not_full;
} queue_t;

// Initialize queue with capacity 'cap'
int queue_init(queue_t *q, size_t cap);
// Destroy queue and free resources
void queue_destroy(queue_t *q);
// Push message (blocking if full)
void queue_push(queue_t *q, message_t msg);
// Try to pop message (non-blocking). Returns 0 on success, -1 if empty
int queue_try_pop(queue_t *q, message_t *msg);
// Pop message (blocking if empty)
message_t queue_pop(queue_t *q);

#endif // QUEUE_H

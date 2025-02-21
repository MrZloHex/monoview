#ifndef __FORWARD_LIST_H__
#define __FORWARD_LIST_H__

/*
 * ----------------------------------------------------------------------------
 *  Forward List Implementation (Macro-Based)
 *
 *  Author  : Zlobin Aleksey
 *  Created : 2025.02.21
 *
 *  Description:
 *    This header defines macros to declare and define type-safe forward lists
 *    in C. Each list stores elements of a particular type T, with generated
 *    function names and a dedicated struct to hold the head pointer and size.
 *
 *  Usage:
 *    1) In your header or C file, use DECLARE_FORWARD_LIST(PREFIX, MT, T) to create
 *       a struct (MT) and function prototypes for your chosen type T.
 *    2) In exactly one C file, include forward_list.h and call
 *       DEFINE_FORWARD_LIST(PREFIX, MT, T) to implement the functions.
 *    3) Use PREFIX##_init, PREFIX##_append, etc. to manipulate your forward list.
 *
 *  Example:
 *    // In mylists.h or mylists.c:
 *    DECLARE_FORWARD_LIST(intlist, IntList, int)
 *
 *    // In mylists.c:
 *    DEFINE_FORWARD_LIST(intlist, IntList, int)
 *
 *    // Now you can do:
 *    IntList list;
 *    intlist_init(&list);
 *    intlist_append(&list, 42);
 *    intlist_remove(&list, 0);
 *    ...
 *
 * ----------------------------------------------------------------------------
 */

#include <stdlib.h>

/*
 * The DECLARE_FORWARD_LIST macro declares a type-safe forward list interface for
 * a specific element type T. It creates a struct (MT) representing the forward list
 * and a set of function prototypes to operate on it. PREFIX is used to build
 * function names uniquely for this element type, ensuring no naming conflicts.
 *
 * Parameters:
 *  - PREFIX: A prefix for the function names, e.g. 'intlist' for int lists.
 *  - MT:     Master Type. The name of the struct type that will represent the forward list.
 *  - T:      The element type that this forward list will hold.
 *
 * All functions with int return type
 *  On success,  0  is returned.
 *  On error,   -1  is returned.
 *
 * Functions with PTR return type
 *  On error,   NULL is returned,
 *    otherwise success.
 */
#define DECLARE_FORWARD_LIST(PREFIX, MT, T)                                \
    typedef struct                                                         \
    {                                                                      \
        T *data;                                                           \
        struct MT *next;                                                   \
    } MT;                                                                  \
                                                                           \
    typedef struct                                                         \
    {                                                                      \
        MT *head;                                                          \
        size_t size;                                                       \
    } PREFIX##_list;                                                       \
                                                                           \
    int                                                                    \
    PREFIX##_init(PREFIX##_list *list);                                    \
                                                                           \
    void                                                                   \
    PREFIX##_deinit(PREFIX##_list *list);                                  \
                                                                           \
    int                                                                    \
    PREFIX##_append(PREFIX##_list *list, T element);                       \
                                                                           \
    int                                                                    \
    PREFIX##_get(const PREFIX##_list *list, size_t index, T *out);         \
                                                                           \
    size_t                                                                 \
    PREFIX##_size(const PREFIX##_list *list);                              \
                                                                           \
    int                                                                    \
    PREFIX##_remove(PREFIX##_list *list, size_t index);


/*
 * The DEFINE_FORWARD_LIST macro defines the functions declared by DECLARE_FORWARD_LIST.
 * It generates implementations for the init, deinit, append, get, size, and remove functions
 * for a given T and associated PREFIX/MT.
 */
#define DEFINE_FORWARD_LIST(PREFIX, MT, T)                                 \
    int                                                                    \
    PREFIX##_init(PREFIX##_list *list)                                     \
    {                                                                      \
        if (!list)                                                         \
        { return -1; }                                                     \
                                                                           \
        list->head = NULL;                                                 \
        list->size = 0;                                                    \
        return 0;                                                          \
    }                                                                      \
                                                                           \
    void                                                                   \
    PREFIX##_deinit(PREFIX##_list *list)                                   \
    {                                                                      \
        if (!list)                                                         \
        { return; }                                                        \
                                                                           \
        MT *current = list->head;                                          \
        MT *next;                                                          \
        while (current)                                                    \
        {                                                                  \
            next = current->next;                                          \
            free(current);                                                 \
            current = next;                                                \
        }                                                                  \
        list->head = NULL;                                                 \
        list->size = 0;                                                    \
    }                                                                      \
                                                                           \
    int                                                                    \
    PREFIX##_append(PREFIX##_list *list, T element)                        \
    {                                                                      \
        if (!list)                                                         \
        { return -1; }                                                     \
                                                                           \
        MT *new_node = (MT *)malloc(sizeof(MT));                           \
        if (!new_node)                                                     \
        { return -1; }                                                     \
                                                                           \
        new_node->data = element;                                          \
        new_node->next = NULL;                                             \
                                                                           \
        if (!list->head)                                                   \
        {                                                                  \
            list->head = new_node;                                         \
        }                                                                  \
        else                                                               \
        {                                                                  \
            MT *last = list->head;                                         \
            while (last->next)                                             \
            {                                                              \
                last = last->next;                                         \
            }                                                              \
            last->next = new_node;                                         \
        }                                                                  \
        list->size++;                                                      \
        return 0;                                                          \
    }                                                                      \
                                                                           \
    int                                                                    \
    PREFIX##_get(const PREFIX##_list *list, size_t index, T *out)          \
    {                                                                      \
        if (!list || !out || index >= list->size)                          \
        { return -1; }                                                     \
                                                                           \
        MT *current = list->head;                                          \
        for (size_t i = 0; i < index; i++)                                 \
        {                                                                  \
            current = current->next;                                       \
        }                                                                  \
        *out = current->data;                                              \
        return 0;                                                          \
    }                                                                      \
                                                                           \
    size_t                                                                 \
    PREFIX##_size(const PREFIX##_list *list)                               \
    {                                                                      \
        return list ? list->size : 0;                                      \
    }                                                                      \
                                                                           \
    int                                                                    \
    PREFIX##_remove(PREFIX##_list *list, size_t index)                     \
    {                                                                      \
        if (!list || index >= list->size)                                  \
        { return -1; }                                                     \
                                                                           \
        MT *current = list->head;                                          \
        MT *prev = NULL;                                                   \
                                                                           \
        if (index == 0)                                                    \
        {                                                                  \
            list->head = current->next;                                    \
        }                                                                  \
        else                                                               \
        {                                                                  \
            for (size_t i = 0; i < index; i++)                             \
            {                                                              \
                prev = current;                                            \
                current = current->next;                                   \
            }                                                              \
            prev->next = current->next;                                    \
        }                                                                  \
        free(current);                                                     \
        list->size--;                                                      \
        return 0;                                                          \
    }

#endif /* __FORWARD_LIST_H__ */


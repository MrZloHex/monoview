#ifndef __DYNARRAY_H__
#define __DYNARRAY_H__

/*
 * ----------------------------------------------------------------------------
 *  Dynamic Array Implementation (Macro-Based)
 *
 *  Author  : Zlobin Aleksey
 *  Created : 2024.12.07
 *
 *  Description:
 *    This header defines macros to declare and define type-safe dynamic arrays
 *    in C. Each array stores elements of a particular type T, with generated
 *    function names and a dedicated struct to hold capacity and size.
 *
 *  Usage:
 *    1) In your header or C file, use DECLARE_DYNARRAY(PREFIX, MT, T) to create
 *       a struct (MT) and function prototypes for your chosen type T.
 *    2) In exactly one C file, include dynarray.h and call
 *       DEFINE_DYNARRAY(PREFIX, MT, T) to implement the functions.
 *    3) Use PREFIX##_init, PREFIX##_append, etc. to manipulate your dynamic array.
 *
 *  Example:
 *    // In myarrays.h or myarrays.c:
 *    DECLARE_DYNARRAY(intarray, IntArray, int)
 *
 *    // In myarrays.c:
 *    DEFINE_DYNARRAY(intarray, IntArray, int)
 *
 *    // Now you can do:
 *    IntArray arr;
 *    intarray_init(&arr, 10);
 *    intarray_append(&arr, 42);
 *    ...
 *
 * ----------------------------------------------------------------------------
 */

#include <stdlib.h>

/*
 * The DECLARE_DYNARRAY macro declares a type-safe dynamic array interface for
 * a specific element type T. It creates a struct (MT) representing the dynamic array
 * and a set of function prototypes to operate on it. PREFIX is used to build
 * function names uniquely for this element type, ensuring no naming conflicts.
 *
 * Parameters:
 *  - PREFIX: A prefix for the function names, e.g. 'intarray' for int arrays.
 *  - MT:     Master Type. The name of the struct type that will represent the dynamic array.
 *  - T:      The element type that this dynamic array will hold.
 *
 *
 * All functions with int return type
 *  On success,  0  is returned.
 *  On error,   -1  is returned.
 *
 * Functions with PTR return type
 *  On error,   NULL is returned,
 *    otherwise success.
 */
#define DECLARE_DYNARRAY(PREFIX, MT, T)                                     \
    typedef struct                                                          \
    {                                                                       \
        T     *data;                                                        \
        size_t size;                                                        \
        size_t capacity;                                                    \
    } MT;                                                                   \
                                                                            \
    int                                                                     \
    PREFIX##_init(MT *arr, size_t initial_capacity);                        \
                                                                            \
    void                                                                    \
    PREFIX##_deinit(MT *arr);                                               \
                                                                            \
    int                                                                     \
    PREFIX##_append(MT *arr, T element);                                    \
                                                                            \
    int                                                                     \
    PREFIX##_get(const MT *arr, size_t index, T *out);                      \
                                                                            \
    T *                                                                     \
    PREFIX##_get_last_ptr(const MT *arr);                                   \
                                                                            \
    size_t                                                                  \
    PREFIX##_size(const MT *arr);                                           \
                                                                            \
    size_t                                                                  \
    PREFIX##_capacity(const MT *arr);   



/*
 * The DEFINE_DYNARRAY macro defines the functions declared by DECLARE_DYNARRAY.
 * It generates implementations for the init, deinit, append, get, get_last_ptr,
 * size, and capacity functions for a given T and associated PREFIX/MT.
 */
#define DEFINE_DYNARRAY(PREFIX, MT, T)                                      \
    int                                                                     \
    PREFIX##_init(MT *arr, size_t initial_capacity)                         \
    {                                                                       \
        if (!arr || initial_capacity == 0)                                  \
        { return -1; }                                                      \
                                                                            \
        arr->data = (T *)malloc(sizeof(T) * initial_capacity);              \
        if (!arr->data)                                                     \
        { return -1; }                                                      \
                                                                            \
        arr->size = 0;                                                      \
        arr->capacity = initial_capacity;                                   \
        return 0;                                                           \
    }                                                                       \
                                                                            \
    void                                                                    \
    PREFIX##_deinit(MT *arr)                                                \
    {                                                                       \
        if (!arr)                                                           \
        { return; }                                                         \
        free(arr->data);                                                    \
        arr->data = NULL;                                                   \
        arr->size = 0;                                                      \
        arr->capacity = 0;                                                  \
    }                                                                       \
                                                                            \
    int                                                                     \
    PREFIX##_append(MT *arr, T element)                                     \
    {                                                                       \
        if (!arr)                                                           \
        { return -1; }                                                      \
                                                                            \
        if (arr->size == arr->capacity)                                     \
        {                                                                   \
            size_t new_capacity = arr->capacity * 2;                        \
            T *temp = (T *)realloc(arr->data, sizeof(T) * new_capacity);    \
            if (!temp)                                                      \
            { return -1; }                                                  \
            arr->data = temp;                                               \
            arr->capacity = new_capacity;                                   \
        }                                                                   \
        arr->data[arr->size++] = element;                                   \
        return 0;                                                           \
    }                                                                       \
                                                                            \
    int                                                                     \
    PREFIX##_get(const MT *arr, size_t index, T *out)                       \
    {                                                                       \
        if (!arr || !out || index >= arr->size)                             \
        { return -1; }                                                      \
        *out = arr->data[index];                                            \
        return 0;                                                           \
    }                                                                       \
                                                                            \
    T *                                                                     \
    PREFIX##_get_last_ptr(const MT *arr)                                    \
    {                                                                       \
        if (!arr)                                                           \
        { return NULL; }                                                    \
        return &arr->data[arr->size - 1];                                   \
    }                                                                       \
                                                                            \
    size_t                                                                  \
    PREFIX##_size(const MT *arr)                                            \
    {                                                                       \
        return arr ? arr->size : 0;                                         \
    }                                                                       \
                                                                            \
    size_t                                                                  \
    PREFIX##_capacity(const MT *arr)                                        \
    {                                                                       \
        return arr ? arr->capacity : 0;                                     \
    }

#endif /* __DYNARRAY_H__ */

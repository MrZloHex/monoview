#ifndef WSCLIENT_H
#define WSCLIENT_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

// WebSocket client structure to hold connection state.
typedef struct {
    struct lws_context *context;
    struct lws *wsi;
    const char *address;
    int port;
    const char *path;
} ws_client_t;

/* 
 * Create a WebSocket client.
 * 'address' - The server address (hostname or IP).
 * 'port'    - The server port (e.g., 80 or 443).
 * 'path'    - The WebSocket path (e.g., "/ws").
 * Returns a pointer to the new ws_client_t or NULL on error.
 */
ws_client_t *ws_client_create(const char *address, int port, const char *path);

/*
 * Connect the WebSocket client to the server.
 * Returns 0 on success or -1 on failure.
 */
int ws_client_connect(ws_client_t *client);

/*
 * Send a message to the WebSocket server.
 * Returns 0 on success or -1 on failure.
 */
int ws_client_send(ws_client_t *client, const char *msg);

/*
 * Service the WebSocket connection (non-blocking).
 * Typically called in an event loop.
 * Returns 0 on success or -1 on failure.
 */
int ws_client_service(ws_client_t *client);

/*
 * Set a callback function to handle incoming messages.
 * 'callback' - The callback function that takes the message as a string.
 */
void ws_client_set_message_callback(void (*callback)(const char *msg));

/*
 * Cleanup and destroy the WebSocket client.
 */
void ws_client_destroy(ws_client_t *client);

#ifdef __cplusplus
}
#endif

#endif // WSCLIENT_H


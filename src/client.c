#include "client.h"
#include <pthread.h>
#include <libwebsockets.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>

// Global variables for WebSocket context and thread control.
static struct lws_context *ws_context = NULL;
static bool ws_running = true;
pthread_t ws_thread;

static int websocket_callback(struct lws *wsi,
                              enum lws_callback_reasons reason,
                              void *user,
                              void *in,
                              size_t len) {
    // Your callback implementation...
    return 0;
}

static struct lws_protocols protocols[] = {
    {
        .name = "my-protocol",
        .callback = websocket_callback,
        .per_session_data_size = 0,
        .rx_buffer_size = 4096,
    },
    { NULL, NULL, 0, 0 }
};

// Declare a static pointer for the client connection.
static struct lws *client_wsi = NULL;

void *ws_service_thread(void *arg) {
    while (ws_running) {
        // Process WebSocket events with a short timeout.
        lws_service(ws_context, 50);
    }
    return NULL;
}

int start_ws_thread(const char *address, int port, const char *path) {
    // Setup your libwebsockets context and connection here.
    // (This example assumes you've already created ws_context and connected.)
    // For example:
    struct lws_context_creation_info info;
    memset(&info, 0, sizeof(info));
    info.port = CONTEXT_PORT_NO_LISTEN;
    info.protocols = protocols;  // your protocol array
    info.options = 0;
    
    ws_context = lws_create_context(&info);
    if (!ws_context) {
        fprintf(stderr, "lws_create_context failed\n");
        return -1;
    }
    
    struct lws_client_connect_info ccinfo = {0};
    ccinfo.context  = ws_context;
    ccinfo.address  = address;
    ccinfo.port     = port;
    ccinfo.path     = path;  // Use "/" if no specific path is required.
    ccinfo.host     = address;
    ccinfo.origin   = address;
    ccinfo.protocol = protocols[0].name;
    
    // Connect using libwebsockets (client_wsi etc.)
    client_wsi = lws_client_connect_via_info(&ccinfo);
    if (!client_wsi) {
        fprintf(stderr, "WebSocket connection failed\n");
        lws_context_destroy(ws_context);
        return -1;
    }
    
    // Create the thread for servicing WebSocket events.
    if (pthread_create(&ws_thread, NULL, ws_service_thread, NULL) != 0) {
        fprintf(stderr, "Failed to create ws thread\n");
        lws_context_destroy(ws_context);
        return -1;
    }
    return 0;
}


void stop_ws_thread() {
    ws_running = false;
    // Cancel the service to break out of lws_service() blocking calls.
    if (ws_context)
        lws_cancel_service(ws_context);
    // Wait for the thread to finish.
    pthread_join(ws_thread, NULL);
    if (ws_context)
        lws_context_destroy(ws_context);
}


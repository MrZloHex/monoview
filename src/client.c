#include "client.h"
#include <libwebsockets.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// Callback for handling WebSocket events.
static void (*message_callback)(const char *msg) = NULL;

static int ws_callback(struct lws *wsi,
                       enum lws_callback_reasons reason,
                       void *user,
                       void *in,
                       size_t len)
{
    ws_client_t *client = (ws_client_t *)user;

    switch (reason) {
        case LWS_CALLBACK_CLIENT_ESTABLISHED:
            printf("WebSocket: Connection established\n");
            break;

        case LWS_CALLBACK_CLIENT_RECEIVE:
            printf("WebSocket: Received: %.*s\n", (int)len, (char *)in);
            if (message_callback) {
                message_callback((const char *)in);
            }
            break;

        case LWS_CALLBACK_CLIENT_WRITEABLE:
            // Handle sending any pending message here if needed
            break;

        case LWS_CALLBACK_CLIENT_CONNECTION_ERROR:
            printf("WebSocket: Connection error\n");
            break;

        case LWS_CALLBACK_CLIENT_CLOSED:
            printf("WebSocket: Connection closed\n");
            break;

        default:
            break;
    }

    return 0;
}

static struct lws_protocols protocols[] = {
    {
        .name = "ws-protocol",
        .callback = ws_callback,
        .per_session_data_size = sizeof(ws_client_t),
        .rx_buffer_size = 4096,
    },
    { NULL, NULL, 0, 0 }
};

ws_client_t *ws_client_create(const char *address, int port, const char *path) {
    ws_client_t *client = (ws_client_t *)malloc(sizeof(ws_client_t));
    if (!client) return NULL;

    client->address = address;
    client->port = port;
    client->path = path;
    client->context = NULL;
    client->wsi = NULL;

    return client;
}

int ws_client_connect(ws_client_t *client) {
    if (!client) return -1;

    struct lws_context_creation_info info;
    memset(&info, 0, sizeof(info));
    info.port = CONTEXT_PORT_NO_LISTEN; // We're a client only.
    info.protocols = protocols;

    client->context = lws_create_context(&info);
    if (!client->context) {
        printf("WebSocket: lws_create_context failed\n");
        return -1;
    }

    struct lws_client_connect_info ccinfo;
    memset(&ccinfo, 0, sizeof(ccinfo));
    ccinfo.context  = client->context;
    ccinfo.address  = client->address;
    ccinfo.port     = client->port;
    ccinfo.path     = client->path;
    ccinfo.host     = client->address;
    ccinfo.origin   = client->address;
    ccinfo.protocol = protocols[0].name;

    client->wsi = lws_client_connect_via_info(&ccinfo);
    if (!client->wsi) {
        printf("WebSocket: Connection failed\n");
        lws_context_destroy(client->context);
        return -1;
    }

    return 0;
}

int ws_client_send(ws_client_t *client, const char *msg) {
    if (!client || !client->wsi) return -1;

    size_t msg_len = strlen(msg);
    unsigned char *buf = malloc(LWS_PRE + msg_len);
    if (!buf) return -1;

    memcpy(&buf[LWS_PRE], msg, msg_len);
    int n = lws_write(client->wsi, &buf[LWS_PRE], msg_len, LWS_WRITE_TEXT);
    free(buf);

    if (n < (int)msg_len) {
        printf("WebSocket: Partial write\n");
        return -1;
    }

    return 0;
}

int ws_client_service(ws_client_t *client) {
    if (!client || !client->context) return -1;

    lws_service(client->context, 0); // Non-blocking service
    return 0;
}

void ws_client_set_message_callback(void (*callback)(const char *msg)) {
    message_callback = callback;
}

void ws_client_destroy(ws_client_t *client) {
    if (client) {
        if (client->context) {
            lws_context_destroy(client->context);
        }
        free(client);
    }
}


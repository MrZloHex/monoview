/* ==============================================================================
 *
 *  ███╗   ███╗ ██████╗ ███╗   ██╗ ██████╗ ██╗     ██╗████████╗██╗  ██╗
 *  ████╗ ████║██╔═══██╗████╗  ██║██╔═══██╗██║     ██║╚══██╔══╝██║  ██║
 *  ██╔████╔██║██║   ██║██╔██╗ ██║██║   ██║██║     ██║   ██║   ███████║
 *  ██║╚██╔╝██║██║   ██║██║╚██╗██║██║   ██║██║     ██║   ██║   ██╔══██║
 *  ██║ ╚═╝ ██║╚██████╔╝██║ ╚████║╚██████╔╝███████╗██║   ██║   ██║  ██║
 *  ╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝  ╚═╝
 *
 *                           ░▒▓█ _MonoView_ █▓▒░
 *
 *   File       : main.c
 *   Author     : MrZloHex
 *   Date       : 2025-02-19
 *
 * ==============================================================================
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <ncurses.h>
#include <sqlite3.h>
#include <libwebsockets.h>
#include <time.h>
#include "queue.h"

#define WS_URL "echo.websocket.org"
#define WS_PORT 443
#define QUEUE_CAP 256
#define GRAPH_WIDTH 50

static queue_t ws_queue;
static queue_t ui_queue;
static queue_t db_queue;
volatile int keep_running = 1;
static sqlite3 *db;
static sqlite3_stmt *insert_stmt;

// WebSocket callback
static int ws_callback(struct lws *wsi, enum lws_callback_reasons reason,
                       void *user, void *in, size_t len) {
    switch (reason) {
    case LWS_CALLBACK_CLIENT_ESTABLISHED:
        lws_callback_on_writable(wsi);
        break;
    case LWS_CALLBACK_CLIENT_WRITEABLE: {
        char buf[64];
        int n = snprintf(buf, sizeof(buf), "{\"rand\":%d}", rand() % 100);
        unsigned char *p = malloc(LWS_PRE + n);
        memcpy(p + LWS_PRE, buf, n);
        lws_write(wsi, p + LWS_PRE, n, LWS_WRITE_TEXT);
        free(p);
        break;
    }
    case LWS_CALLBACK_CLIENT_RECEIVE: {
        message_t m;
        m.payload = malloc(len+1);
        memcpy(m.payload, in, len);
        ((char*)m.payload)[len] = '\0';
        queue_push(&ws_queue, m);
        break;
    }
    case LWS_CALLBACK_CLOSED:
        keep_running = 0;
        break;
    default:
        break;
    }
    return 0;
}

void *ws_thread(void *arg) {
    struct lws_context_creation_info info = {0};
    info.port = CONTEXT_PORT_NO_LISTEN;
    info.protocols = (struct lws_protocols[]){{"chat", ws_callback, 0, 0}, {NULL, NULL, 0,0}};
    struct lws_context *ctx = lws_create_context(&info);
    if (!ctx) return NULL;
    struct lws_client_connect_info cc = {0};
    cc.context = ctx;
    cc.address = WS_URL;
    cc.port = WS_PORT;
    cc.path = "/";
    cc.host = lws_canonical_hostname(ctx);
    cc.origin = "origin";
    cc.ssl_connection = LCCSCF_USE_SSL;
    cc.protocol = "chat";
    if (!lws_client_connect_via_info(&cc)) { lws_context_destroy(ctx); return NULL; }
    while (keep_running) lws_service(ctx, 100);
    lws_context_destroy(ctx);
    return NULL;
}

void *processor_thread(void *arg) {
    while (keep_running) {
        message_t m = queue_pop(&ws_queue);
        int value;
        if (sscanf((char*)m.payload, "{\"rand\":%d}", &value)==1) {
            int *v = malloc(sizeof(int)); *v = value;
            queue_push(&ui_queue, (message_t){.payload=v});
            time_t t = time(NULL);
            char *log = malloc(64);
            snprintf(log,64, "%ld,%d", t, value);
            queue_push(&db_queue, (message_t){.payload=log});
        }
        free(m.payload);
    }
    return NULL;
}


void *db_thread(void *arg) {
    if (sqlite3_open("local.db", &db)!=SQLITE_OK) return NULL;
    sqlite3_exec(db, "CREATE TABLE IF NOT EXISTS logs(ts INTEGER, val INTEGER);", NULL,NULL,NULL);
    sqlite3_prepare_v2(db, "INSERT INTO logs VALUES(?,?);", -1, &insert_stmt, NULL);
    while (keep_running) {
        message_t m = queue_pop(&db_queue);
        long ts; int v;
        sscanf((char*)m.payload, "%ld,%d", &ts, &v);
        sqlite3_bind_int64(insert_stmt,1,ts);
        sqlite3_bind_int(insert_stmt,2,v);
        sqlite3_step(insert_stmt);
        sqlite3_reset(insert_stmt);
        free(m.payload);
    }
    sqlite3_finalize(insert_stmt);
    sqlite3_close(db);
    return NULL;
}

#include "ui/ui.h"

int main(void) {
//    pthread_t ws_thr, proc_thr, db_thr;
//    queue_init(&ws_queue, QUEUE_CAP);
    queue_init(&ui_queue, QUEUE_CAP);
//    queue_init(&db_queue, QUEUE_CAP);
//    pthread_create(&ws_thr,NULL,ws_thread,NULL);
//    pthread_create(&proc_thr,NULL,processor_thread,NULL);
//    pthread_create(&db_thr,NULL,db_thread,NULL);
    module_ui_entry(&ui_queue);
    keep_running=0;
//    pthread_join(ws_thr,NULL);
//    pthread_join(proc_thr,NULL);
//    pthread_join(db_thr,NULL);
    queue_destroy(&ws_queue);
    queue_destroy(&ui_queue);
    queue_destroy(&db_queue);
    return 0;
}


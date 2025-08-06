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

#include <locale.h>
#include <stdlib.h>
#include <unistd.h>
#include "trace.h"

#include "ui/ui.h"
#include "logic/logic.h"

int
main(void)
{
    setlocale(LC_ALL, "");
    FILE *log = fopen("log", "a");
    tracer_init(TRC_DEBUG, TP_ALL, -1);
    tracer_add_stream(log);

    TRACE_INFO(" =========== FULL START ===========");

    pthread_t ui_trd, logic_trd;
    if (pthread_create(&ui_trd, NULL, ui_thread, NULL))
    {
        TRACE_FATAL("Failed to init UI THREAD");
        exit(EXIT_FAILURE);
    }
    if (pthread_create(&logic_trd, NULL, logic_thread, NULL))
    {
        TRACE_FATAL("Failed to init LOGIC THREAD");
        exit(EXIT_FAILURE);
    }


    if (pthread_join(ui_trd, NULL))
    {
        TRACE_FATAL("Failed to join UI THREAD");
        exit(EXIT_FAILURE);
    }
    pthread_cancel(logic_trd);

    TRACE_INFO(" ============ FULL STOP ============== ");

    usleep(1000);
    tracer_stop();
    fclose(log);
    return 0;
}


#ifndef __CLIENT_H__
#define __CLIENT_H__

int start_ws_thread(const char *address, int port, const char *path);


void stop_ws_thread();


#endif /* __CLIENT_H__ */

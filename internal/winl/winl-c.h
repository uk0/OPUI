//
//  hx-win.h
//  WinGL
//
//  Created by 陈成涛 on 09/09/2017.
//  Copyright © 2017 chenct. All rights reserved.
//

#ifndef winl_c_h
#define winl_c_h

#include <stdint.h>

#ifdef __linux__
  typedef int NativeWnd;
#else
  typedef void* NativeWnd;
#endif

enum {
  WINL_HINT_RESIZABLE = 0x0002,
  WINL_HINT_ANIMATE   = 0x0004,
  WINL_HINT_PAINTER   = 0x0008,
  WINL_HINT_3D        = 0x0010,
};

enum {
	WINL_MOUSE_BTN_LEFT = 1,
	WINL_MOUSE_BTN_RIGHT = 2,
  WINL_MOUSE_BTN_MIDDLE = 4,
};

void winl_get_screen_size(int *width, int *height);

NativeWnd winl_create(int ws, int width, int height);
void winl_show(NativeWnd win);
int winl_is_visible(NativeWnd win);
void winl_destroy(NativeWnd win);
void winl_set_title(NativeWnd win, const char * title);
void winl_get_size(NativeWnd win, float *width, float *height);
int winl_is_full_screen(NativeWnd win);
void winl_toggle_full_screen(NativeWnd win);
int winl_make_current(NativeWnd win); // pass 0 to release current context
void winl_swap_buffers(NativeWnd win);
int winl_event_loop();
void winl_exit_loop(int code);
char* winl_os_version(); // use free to release memory
void winl_expose(NativeWnd win, float x, float y, float width, float height);

// event handlers is implement in winl.go
extern void winl_on_start();
extern void winl_on_exit(int code);
extern void winl_on_destroy(NativeWnd win);
extern void winl_on_resize(NativeWnd win, float width, float height);
extern void winl_on_mouse_move(NativeWnd win, float x, float y);
extern void winl_on_mouse_press(NativeWnd win, int btn, float x, float y);
extern void winl_on_mouse_release(NativeWnd win, int btn, float x, float y);
extern void winl_on_mouse_wheel(NativeWnd win, int vertical, float dz);
extern void winl_on_mouse_enter(NativeWnd win, float x, float y);
extern void winl_on_mouse_leave(NativeWnd win, float x, float y);
extern void winl_on_expose(NativeWnd win, float x, float y, float width, float height);

extern void winl_report(char* msg, int panic);

void winl_message_box(NativeWnd win, const char* title, const char* msg);
int winl_confirm_box(NativeWnd win, const char* title, const char* msg);

int winl_printf(const char *fmt, ...);
int winl_panicf(const char *fmt, ...);

#endif /* winl_c_h */

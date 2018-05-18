#include <windows.h>
#include <windowsx.h>
#include <stdio.h>
//#include <assert.h>
#include <GL/GL.h>
#include "winl-c.h"


#ifndef WM_MOUSEHWHEEL
#	define WM_MOUSEHWHEEL 0x020E
#endif

#define szOpenGLWndClass "WINL_OPENGL"

LRESULT CALLBACK MainWndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam);
LRESULT CALLBACK OpenGLWndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam);
void InitOpenGL(HWND hWnd);
//void CleanUp(HWND hWnd);

HGLRC _hGLRC;
int _windowCount;

typedef struct NativeWndData {
	int trackMouse : 1;
	int mouseHover : 1;
	int fullScreen: 1;
	RECT restoreRect;
	DWORD restoreStyle;
	DWORD restoreExStyle;
	int btnDown;
	HDC hDC;
}NativeWndData;

static NativeWndData* getWndData(HWND hWnd) {
	return (NativeWndData*)(void*)GetWindowLongPtr(hWnd, GWLP_USERDATA);
}


void MyRegisterClass()
{
	static BOOL _inited;
	WNDCLASSEX wcex;

	if(_inited)
		return;
	_inited = TRUE;

	HINSTANCE hInstance = GetModuleHandle(NULL);

	wcex.cbSize = sizeof(WNDCLASSEX);

	wcex.style			= CS_HREDRAW | CS_VREDRAW | CS_OWNDC;
	wcex.lpfnWndProc	= OpenGLWndProc;
	wcex.cbClsExtra		= 0;
	wcex.cbWndExtra		= 0;
	wcex.hInstance		= hInstance;
	wcex.hIcon			= NULL; // LoadIcon(hInstance, MAKEINTRESOURCE(IDI_TESTGL));
	wcex.hCursor		= LoadCursor(NULL, IDC_ARROW);
	wcex.hbrBackground	= NULL;
	wcex.lpszMenuName	= NULL;
	wcex.lpszClassName	= szOpenGLWndClass;
	wcex.hIconSm		= NULL;//LoadIcon(wcex.hInstance, MAKEINTRESOURCE(IDI_SMALL));

	RegisterClassEx(&wcex);

	return;
}

static void get_mouse_pos(NativeWnd win, float *x, float *y) {

	POINT pt;
	if(win) {
		GetCursorPos(&pt);
		ScreenToClient((HWND)win, &pt);
	} else {
		pt.x = 0; pt.y = 0;
	}
	if(x) {
		*x = (float)pt.x;
	}
	if(y) {
		*y = (float)pt.y;
	}
}

LRESULT CALLBACK OpenGLWndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam) {
	PAINTSTRUCT ps;
	NativeWndData* wd = getWndData(hWnd);
	switch (message) {
	case WM_CREATE: {

		wd = malloc(sizeof(NativeWndData));
		memset(wd, 0, sizeof(NativeWndData));
		wd->hDC = GetDC(hWnd);
		SetWindowLongPtr(hWnd, GWLP_USERDATA, (UINT_PTR)wd);
		InitOpenGL(hWnd);
	} break; case WM_TIMER: {
		//if(wParam == nTimerRedraw)
		//{
		//	OnRedrawTimer(hWnd);
		//}
	} break; case WM_LBUTTONDOWN: {
		if(wd->trackMouse && !wd->btnDown) {
			SetCapture(hWnd);
		}
		wd->btnDown |= WINL_MOUSE_BTN_LEFT;
		winl_on_mouse_press(hWnd, WINL_MOUSE_BTN_LEFT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
	} break; case WM_LBUTTONUP: {
		wd->btnDown &= ~WINL_MOUSE_BTN_LEFT;
		winl_on_mouse_release(hWnd, WINL_MOUSE_BTN_LEFT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		if(wd->trackMouse && !wd->btnDown) {
			ReleaseCapture();
		}
	} break; case WM_RBUTTONDOWN: {
		if(wd->trackMouse && !wd->btnDown) {
			SetCapture(hWnd);
		}
		wd->btnDown |= WINL_MOUSE_BTN_RIGHT;
		winl_on_mouse_press(hWnd, WINL_MOUSE_BTN_RIGHT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
	} break; case WM_RBUTTONUP: {
		wd->btnDown &= ~WINL_MOUSE_BTN_RIGHT;
		winl_on_mouse_release(hWnd, WINL_MOUSE_BTN_RIGHT, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		if(wd->trackMouse && !wd->btnDown) {
			ReleaseCapture();
		}
	} break; case WM_MOUSEMOVE: {
		if(wd->trackMouse && !wd->mouseHover) {
			wd->mouseHover = 1;
			TRACKMOUSEEVENT tme = { sizeof(TRACKMOUSEEVENT), TME_LEAVE, hWnd, HOVER_DEFAULT};
			TrackMouseEvent(&tme); // one shot
			winl_on_mouse_enter(hWnd, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
		}
		winl_on_mouse_move(hWnd, GET_X_LPARAM(lParam), GET_Y_LPARAM(lParam));
	} break; case WM_MOUSEWHEEL: {
		winl_on_mouse_wheel(hWnd, 1, GET_WHEEL_DELTA_WPARAM(wParam));
	} break; case WM_MOUSEHWHEEL: {
		winl_on_mouse_wheel(hWnd, 0, GET_WHEEL_DELTA_WPARAM(wParam));
	} break; case WM_MOUSELEAVE: {
		if(wd->trackMouse && wd->mouseHover) {
			wd->mouseHover = 0;
			float x, y;
			get_mouse_pos(hWnd, &x, &y);
			winl_on_mouse_leave(hWnd, x, y);
		}
	} break; case WM_PAINT: {
		HDC hdc = BeginPaint(hWnd, &ps);
		RECT rc = ps.rcPaint;
		winl_on_expose(hWnd, (float)(rc.left), (float)(rc.top), (float)(rc.right - rc.left), (float)(rc.bottom - rc.top));
		EndPaint(hWnd, &ps);
	} break; case WM_SIZE: {
		winl_on_resize(hWnd, LOWORD(lParam), HIWORD(lParam));
	} break; case WM_DESTROY: {
		winl_on_destroy(hWnd);
		_windowCount--;
		ReleaseDC(hWnd, wd->hDC);
		free(wd);
	} break; default: {
		return DefWindowProc(hWnd, message, wParam, lParam);
	}}
	return 0;
}

//GLint maxClipPlanes = 0;

void InitOpenGL(HWND hWnd)
{
		winl_printf("InitOpenGL\n");

	NativeWndData* wd = getWndData(hWnd);

	static PIXELFORMATDESCRIPTOR pfd =				// pfd Tells Windows How We Want Things To Be
	{
		sizeof(PIXELFORMATDESCRIPTOR),				// Size Of This Pixel Format Descriptor
		1,											// Version Number
		PFD_DRAW_TO_WINDOW | PFD_SUPPORT_OPENGL | PFD_DOUBLEBUFFER ,
		PFD_TYPE_RGBA,								// Request An RGBA Format
		24,											// Select Our Color Depth
		0, 0, 0, 0, 0, 0,							// Color Bits Ignored
		0,											// No Alpha Buffer
		0,											// Shift Bit Ignored
		0,											// No Accumulation Buffer
		0, 0, 0, 0,									// Accumulation Bits Ignored
		0,											// 16Bit Z-Buffer (Depth Buffer)
		0,											// No Stencil Buffer
		0,											// No Auxiliary Buffer
		PFD_MAIN_PLANE,								// Main Drawing Layer
		0,											// Reserved
		0, 0, 0										// Layer Masks Ignored
	};

	//HDC hDC = GetDC(hWnd);
	GLuint pixelFormat;
	if (!(pixelFormat = ChoosePixelFormat(wd->hDC, &pfd)))	// D id Windows Find A Matching Pixel Format?
	{
		winl_panicf("Bad pixel format");
		return;
	}

	if (!SetPixelFormat(wd->hDC, pixelFormat, &pfd))		// Are We Able To Set The Pixel Format?
	{
		winl_panicf("Bad pixel format");
		return;
	}


	// HGLRC �ǹ�����, ֻ����һ��
	if (_hGLRC == NULL && !(_hGLRC = wglCreateContext(wd->hDC)))
	{
		winl_panicf("Failed to create opengl context"); // ���� OpenGL ʧ��
		return;
	}


	if (!wglMakeCurrent(wd->hDC, _hGLRC))					// Try To Activate The Rendering Context
	{
		winl_panicf("Failed to activate opengl context");
		return;
	}
	//~ MoveWindow(hWnd, 0, 0, 400, 400, TRUE);
	//~ ShowWindow(hWnd, SW_SHOW);
	//~ glViewport(0,0,300,300);
		//~ glColor4f(1,0.5,0.5,1);
		//~ glClear(GL_COLOR_BUFFER_BIT);
		//~ glFlush();
		//~ SwapBuffers(wd->hDC);
	//if (!GLeeInit())
	//{
	//	MessageBoxA(hWnd, GLeeGetErrorString(), NULL, MB_OK);
	//	return;
	//}
/*
	glGetIntegerv(GL_MAX_CLIP_PLANES, &maxClipPlanes);
	assert(maxClipPlanes >= 4);

	glEnable(GL_BLEND);
	assert(CheckLastError());
	glBlendFunc(GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA);
	assert(CheckLastError());
	glShadeModel(GL_SMOOTH);							// Enable Smooth Shading
	assert(CheckLastError());
	glEnable(GL_LINE_SMOOTH);
	assert(CheckLastError());
	glHint(GL_LINE_SMOOTH_HINT, GL_NICEST);
	assert(CheckLastError());
	glEnable(GL_POINT_SMOOTH);
	assert(CheckLastError());
	glHint(GL_POINT_SMOOTH_HINT, GL_NICEST);
	assert(CheckLastError());
	glEnable(GL_POLYGON_SMOOTH);
	assert(CheckLastError());
    glHint(GL_POLYGON_SMOOTH_HINT, GL_NICEST);
	assert(CheckLastError());
	glClearColor(0.7f, 0.7f, 0.7f, 1.0f);				// Black Background
	assert(CheckLastError());
	glClearDepth(1.0f);									// Depth Buffer Setup
	assert(CheckLastError());


	glGenFramebuffersEXT(1, &fbo);
	assert(CheckLastError());
*/
//	ResizeOpenGL(hWnd);
}

static BOOL pumpMessage(MSG *msg) {
	if(PeekMessage(msg, NULL, 0, 0, PM_REMOVE)) {
		if(msg->message != WM_QUIT) {
			TranslateMessage(msg);
			DispatchMessage(msg);
		}
		return TRUE;
	} else {
		return FALSE;
	}
}

int winl_event_loop(){
	winl_printf("winl_event_loop\n");
  MSG msg;
	int exitCode = 0;

	winl_on_start();
	while(_windowCount > 1) {
		if(!pumpMessage(&msg)) {
			Sleep(1);
		} else if(msg.message == WM_QUIT) {
			exitCode = (int)msg.wParam;
			break;
		}
	}
	winl_on_exit(exitCode);

	return exitCode;
}

void winl_get_screen_size(int *width, int *height) {
	if (width) {
		*width = GetSystemMetrics(SM_CXSCREEN);
	}
	if (height) {
		*height = GetSystemMetrics(SM_CYSCREEN);
	}
}

NativeWnd winl_create(int ws, int width, int height) {
	MyRegisterClass();

	int sw, sh;
	winl_get_screen_size(&sw, &sh);
	if (width <= 0) {
		width = sw / 2;
	}
	if (height <= 0) {
		height = sh / 2;
	}

	int x = (sw - width) / 2;
	int y = (sh - height) / 2;

	DWORD style = WS_OVERLAPPEDWINDOW;
	if((ws & WINL_HINT_RESIZABLE) == 0) {
		style &= ~(WS_THICKFRAME | WS_MAXIMIZEBOX);
	}
	RECT rect = {x, y, x+width, y+height};
	AdjustWindowRect(&rect, style, FALSE);
	HWND hWnd = CreateWindow(szOpenGLWndClass, "", style,
		rect.left, rect.top, rect.right-rect.left, rect.bottom-rect.top,
		NULL, NULL, GetModuleHandle(NULL), NULL);
	if (hWnd) {
		_windowCount++;
	}
   return hWnd;
}

void winl_show(NativeWnd win) {
	if(win) {
		ShowWindow((HWND)win, SW_SHOW);
	}
}

void winl_destroy(NativeWnd win) {
	if(win) {
		DestroyWindow((HWND)win);
	}
}

void winl_track_mouse(NativeWnd win, int enable) {
	if(win == 0) {
		return;
	}
	NativeWndData* wd = getWndData((HWND)win);
	enable = enable ? 1 : 0;
	if(wd->trackMouse != enable) {
		wd->trackMouse = enable;
		if(!enable) {
			if(wd->btnDown) {
				ReleaseCapture();
				wd->btnDown = 0;
			}
		}
		if(enable) {
			wd->mouseHover = 0;
		}
	}
}

int winl_make_current(NativeWnd win) {
	winl_printf("winl_make_current %ld", win);
	if(win) {
		return wglMakeCurrent(getWndData((HWND)win)->hDC, _hGLRC);
	} else {
		return wglMakeCurrent(NULL, NULL);
	}
}

void winl_swap_buffers(NativeWnd win) {
	winl_printf("winl_swap_buffers %ld", win);
	if(win) {
		SwapBuffers(getWndData((HWND)win)->hDC);
	}
}

int winl_is_full_screen(NativeWnd win) {
	if(!win) {
		return 0;
	}
	return getWndData((HWND)win)->fullScreen;
}

int winl_is_visible(NativeWnd win) {
	if(!win) {
		return 0;
	}
	return IsWindowVisible((HWND)win);
}


void winl_toggle_full_screen(NativeWnd win) {
	if(!win) {
		return;
	}
	HWND hWnd = (HWND)win;
	NativeWndData* wd = getWndData(hWnd);
	if(wd->fullScreen) {
		wd->fullScreen = 0;
		SetWindowLongPtr(hWnd, GWL_STYLE, (UINT_PTR)wd->restoreStyle);
		SetWindowLongPtr(hWnd, GWL_EXSTYLE, (UINT_PTR)wd->restoreExStyle);
		RECT rc = wd->restoreRect;
		SetWindowPos(hWnd, NULL, rc.left, rc.top, rc.right-rc.left, rc.bottom-rc.top, SWP_NOZORDER);
	} else {
		wd->fullScreen = 1;
		GetWindowRect(hWnd, &wd->restoreRect);
		wd->restoreStyle = (DWORD)GetWindowLongPtr(hWnd, GWL_STYLE);
		wd->restoreExStyle = (DWORD)GetWindowLongPtr(hWnd, GWL_EXSTYLE);
		DWORD ws = (wd->restoreStyle & ~(WS_CAPTION | WS_THICKFRAME | WS_BORDER)) |
			WS_MAXIMIZEBOX | WS_MAXIMIZE;
		DWORD wsex = wd->restoreExStyle | WS_EX_TOPMOST;
		SetWindowLongPtr(hWnd, GWL_STYLE, (UINT_PTR)ws);
		SetWindowLongPtr(hWnd, GWL_EXSTYLE, (UINT_PTR)wsex);
		int sw, sh;
		winl_get_screen_size(&sw, &sh);
		SetWindowPos(hWnd, NULL, 0, 0, sw, sh, SWP_NOZORDER);
	}
}

void winl_set_title(NativeWnd win, const char * title) {
	if(!win) {
		return;
	}
	int sz = MultiByteToWideChar(CP_UTF8, 0, title, -1, NULL, 0);
	WCHAR* buf = malloc(sizeof(WCHAR)*sz);
	MultiByteToWideChar(CP_UTF8, 0, title, -1, buf, sz);
	SetWindowTextW((HWND)win, buf);
	free(buf);
}

void winl_expose(NativeWnd win, float x, float y, float width, float height) {
	if(!win) {
		return;
	}
	RECT rc = { (int)x, (int)y, (int)(x + width), (int)(y + height) };
	InvalidateRect((HWND)win, &rc, FALSE);
}

void winl_exit_loop(int code) {
	PostQuitMessage(code);
}

static int messasge_box(HWND hWnd,  const char* msg, const char* title, UINT uTypes) {
	int sz;

	sz = MultiByteToWideChar(CP_UTF8, 0, msg, -1, NULL, 0);
	WCHAR* ms = malloc(sizeof(WCHAR)*sz);
	MultiByteToWideChar(CP_UTF8, 0, msg, -1, ms, sz);

	sz = MultiByteToWideChar(CP_UTF8, 0, title, -1, NULL, 0);
	WCHAR* ts = malloc(sizeof(WCHAR)*sz);
	MultiByteToWideChar(CP_UTF8, 0, title, -1, ts, sz);

	int ret = MessageBoxW(hWnd, ms, ts, uTypes);

	free(ms);
	free(ts);

	return ret;
}

void winl_message_box(NativeWnd win, const char* msg, const char* title) {
	messasge_box((HWND)win, msg, title, MB_OK);
}

int winl_confirm_box(NativeWnd win, const char* msg, const char* title) {
	if(IDOK == messasge_box((HWND)win, msg, title, MB_OKCANCEL|MB_ICONQUESTION)) {
		return 1;
	} else {
		return 0;
	}
}
#ifndef VER_MAJORVERSION
#	define VER_MAJORVERSION 0x0000002
#	define VER_MINORVERSION 0x0000001
#	define VER_GREATER_EQUAL 3
#	define VER_SET_CONDITION(_m_,_t_,_c_)  \
        ((_m_)=VerSetConditionMask((_m_),(_t_),(_c_)))

ULONGLONG NTAPI VerSetConditionMask(ULONGLONG ConditionMask, DWORD TypeMask, BYTE  Condition);
#endif

static BOOL windows_aleast(DWORD major, DWORD minor) {
	OSVERSIONINFOEX info;
	memset(&info, 0, sizeof(info));
	info.dwOSVersionInfoSize = sizeof(OSVERSIONINFOEX);
	info.dwMajorVersion = major;
	info.dwMinorVersion = minor;
	DWORDLONG comparison = 0;
	VER_SET_CONDITION(comparison, VER_MAJORVERSION, VER_GREATER_EQUAL);
	VER_SET_CONDITION(comparison, VER_MINORVERSION, VER_GREATER_EQUAL);
	return VerifyVersionInfo(&info, VER_MAJORVERSION | VER_MINORVERSION, comparison);
}

typedef struct VersionItem {
	const char * str;
	DWORD major;
	DWORD minor;
}VersionItem;


char* winl_os_version() {
	VersionItem list[] = {
		{"Windows 10, Windows Server 2016, or Later (%ld.%ld)",	10, 0},
		{"Windows 8.1, Windows Server 2012 R2 (%ld.%ld)",	6, 3},
		{"Windows 8, Windows Server 2012 (%ld.%ld)",	6, 2},
		{"Windows 7, Windows Server 2008 R2 (%ld.%ld)",	6, 1},
		{"Windows Vista, Windows Server 2008 (%ld.%ld)",	6,0},
		{"Windows Server 2003 (%ld.%ld)",	5, 2},
		{"Windows XP (%ld.%ld)",	5, 1},
		{"Windows 2000 (%ld.%ld)",	5, 0}
	};
	VersionItem* item = NULL;
	int i;
	for(i=0; i<sizeof(list)/sizeof(VersionItem); i++) {
		if (windows_aleast(list[i].major, list[i].minor)) {
			item = list + i;
			break;
		}
	}
  char* buf = (char*)malloc(256);
  if (item == NULL) {
		buf[0] = 0;
		strcat(buf, "Windows 9x");
	}else {
		sprintf(buf, item->str, item->major, item->minor);
	}
  return buf;
}

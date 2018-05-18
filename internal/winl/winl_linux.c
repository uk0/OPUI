#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <errno.h>
#include <string.h>
#include <sys/utsname.h>
#include <X11/Xatom.h>
#include <X11/Xlib.h>
#include <GL/gl.h>
#include <GL/glu.h>
#define GLX_GLXEXT_LEGACY
#include <GL/glx.h>
#include "winl-c.h"


static XIM _xim = 0;
static XIC _xic = 0;
Display * _display = 0;
int _screenNum = 0;
Colormap _screenColormap;
//	typedef ::Window WindowID;
Atom _atom_WM_PROTOCOLS;
Atom _atom_WM_DELETE_WINDOW;
//Atom _atom_E_WM_PENDING_CALL;
//Atom _atom_E_WM_CLOSE;
Atom _atom_UTF8_STRING;
Atom _atom_CLIPBOARD;
//Atom _atom_E_SELECTION;
Atom _atom_TARGETS;
Atom _atom_NET_WM_NAME;
Atom _atom_NET_WM_PING;
Atom _atom_NET_WM_ICON;
Atom _atom_NET_WM_STATE;
Atom _atom_NET_WM_STATE_MODAL;
//Atom _atom_NET_WM_STATE_NORMAL;
//Atom _atom_NET_WM_STATE_ADD;
//Atom _atom_NET_WM_STATE_REMOVE;
Atom _atom_NET_WM_STATE_MAXIMIZED_VERT;
Atom _atom_NET_WM_STATE_MAXIMIZED_HORZ;
Atom _atom_NET_WM_STATE_ABOVE;
Atom _atom_NET_WM_STATE_SKIP_TASKBAR;
Atom _atom_NET_WM_STATE_SKIP_PAGER;
Atom _atom_NET_WM_STATE_FOCUSED;
Atom _atom_NET_WM_STATE_SHADED;
Atom _atom_NET_WM_STATE_HIDDEN;
Atom _atom_NET_WM_STATE_FULLSCREEN;

XContext _wdContext;

GLXContext hGL;
int _windowCount;
int _toExit;
int _exitCode;

#ifndef MIN
# define MIN(x, y)  ((x) < (y) ? (x) : (y))
# define MAX(x, y)  ((x) > (y) ? (x) : (y))
#endif

static Atom getAtom(const char* name) {
  Atom atom = XInternAtom(_display, name, True);
  if(atom == None) {
    winl_printf("bad atom: \"%s\"\n", name);
  }
  return atom;
}

static Atom newAtom(const char* name) {
  return XInternAtom(_display, name, False);
}

static void _InitXLib()
{
  if(_display != 0) {
    return;
  }
  if((_display = XOpenDisplay(0)) == 0)
  {
    winl_printf("%s\n", "fatal error: failed to open display.");
    exit(1);
  }
  _screenNum = DefaultScreen(_display);
  _screenColormap = DefaultColormap(_display, DefaultScreen(_display));

  //XSetLocaleModifiers("@im=SCIM");
  XSetLocaleModifiers("");
  _xim = XOpenIM(_display, NULL, NULL, NULL);

  if(_xim)
  {
    Bool fl_is_over_the_spot = False;
    XFontSet fontSet = NULL;
    XRectangle status_area;
    {
      char **missing_list;
      int missing_count;
      char *def_string;
      fontSet = XCreateFontSet(_display, "-misc-fixed-medium-r-normal--14-*", &missing_list, &missing_count, &def_string);
    }

    XIMStyles* xim_styles = NULL;
    if(XGetIMValues(_xim, XNQueryInputStyle, &xim_styles, NULL, NULL) || !xim_styles || !xim_styles->count_styles)
    {
      //hx_warning("No XIM style found\n");
      //assert(0);
      winl_printf("%s\n", "warning: XGetIMValues() failed.");
    }
    else
    {
      //hx_trace("[e_ui] 1111111111111111.");
      Bool predit = False;
      Bool statusArea = False;
      for(int i = 0; i < xim_styles->count_styles; ++i)
      {
        XIMStyle* style = xim_styles->supported_styles+i;
        if(*style == (XIMPreeditPosition | XIMStatusArea))
        {
          statusArea = True;
          predit = True;
        }
        else if(*style == (XIMPreeditPosition | XIMStatusNothing))
        {
          predit = True;
        }
      }

      XFree(xim_styles);

      if(predit)
      {
        //hx_trace("[e_ui] XCreateIC() predit == True.");
        XPoint	spot;
        spot.x = 0;
        spot.y = 0;
        XVaNestedList preedit_attr = XVaCreateNestedList(0, XNSpotLocation, &spot, XNFontSet, fontSet, NULL);

        if(statusArea)
        {
          //hx_trace("[e_ui] XCreateIC() statusArea.");
          fprintf(stdout, "%s\n", "XCreateIC() statusArea.");
          XVaNestedList status_attr = XVaCreateNestedList(0,
              XNAreaNeeded, &status_area,
              XNFontSet, fontSet, NULL);
          _xic = XCreateIC(_xim,
              XNInputStyle, (XIMPreeditPosition | XIMStatusArea),
              XNPreeditAttributes, preedit_attr,
              XNStatusAttributes, status_attr,
              NULL);
          XFree(status_attr);
        }
        _xic = XCreateIC(_xim, XNInputStyle,XIMPreeditPosition | XIMStatusNothing, XNPreeditAttributes, preedit_attr, NULL);
        XFree(preedit_attr);
        if(_xic)
        {
          fprintf(stdout, "%s\n", "XCreateIC() predit.");
          fl_is_over_the_spot = True;
          XVaNestedList status_attr;
          status_attr = XVaCreateNestedList(0, XNAreaNeeded, &status_area, NULL);
          if(status_area.height != 0)
          {
            XGetICValues(_xic, XNStatusAttributes, status_attr, NULL);
          }
          XFree(status_attr);
        }
      }
    }

    if(_xic == 0)
    {
      _xic = XCreateIC(_xim, XNInputStyle, XIMPreeditNothing | XIMStatusNothing, NULL);
    }
    if(_xic == 0)
    {
      //assert(0);
      winl_printf("%s\n", "error: XCreateIC() failed.");
    }
  }
  else
  {
    //assert(0);
    winl_printf("%s\n", "error: XOpenIM() failed.");
  }

  _atom_WM_DELETE_WINDOW  = getAtom("WM_DELETE_WINDOW");
  //_atom_E_WM_PENDING_CALL = getAtom("E_WM_PENDING_CALL");
  //_atom_E_WM_CLOSE        = getAtom("E_WM_CLOSE");
  _atom_WM_PROTOCOLS      = getAtom("WM_PROTOCOLS");
  _atom_UTF8_STRING       = getAtom("UTF8_STRING");
  _atom_CLIPBOARD         = getAtom("CLIPBOARD");
//  _atom_E_SELECTION       = getAtom("E_SELECTION");
  _atom_TARGETS           = getAtom("TARGETS");
  _atom_NET_WM_NAME       = getAtom("_NET_WM_NAME");
  _atom_NET_WM_PING       = getAtom("_NET_WM_PING");
  _atom_NET_WM_ICON       = getAtom("_NET_WM_ICON");
  _atom_NET_WM_STATE      = getAtom("_NET_WM_STATE");
  _atom_NET_WM_STATE_MODAL = getAtom("_NET_WM_STATE_MODAL");
  //_atom_NET_WM_STATE_NORMAL= getAtom("_NET_WM_STATE_NORMAL");
  //_atom_NET_WM_STATE_ADD   = getAtom("_NET_WM_STATE_ADD");
  //_atom_NET_WM_STATE_REMOVE= getAtom("_NET_WM_STATE_REMOVE");
  _atom_NET_WM_STATE_ABOVE = getAtom("_NET_WM_STATE_ABOVE");
  _atom_NET_WM_STATE_SKIP_TASKBAR = getAtom("_NET_WM_STATE_SKIP_TASKBAR");
  _atom_NET_WM_STATE_SKIP_PAGER = getAtom("_NET_WM_STATE_SKIP_PAGER");
  _atom_NET_WM_STATE_FOCUSED = getAtom("_NET_WM_STATE_FOCUSED");
  _atom_NET_WM_STATE_MAXIMIZED_VERT = getAtom("_NET_WM_STATE_MAXIMIZED_VERT");
  _atom_NET_WM_STATE_MAXIMIZED_HORZ = getAtom("_NET_WM_STATE_MAXIMIZED_HORZ");
  _atom_NET_WM_STATE_SHADED = getAtom("_NET_WM_STATE_SHADED");
  _atom_NET_WM_STATE_HIDDEN = getAtom("_NET_WM_STATE_HIDDEN");
  _atom_NET_WM_STATE_FULLSCREEN = getAtom("_NET_WM_STATE_FULLSCREEN");
  _wdContext = XUniqueContext();
}

static void _CloseXLib()
{
  if(_xic)
  {
    XDestroyIC(_xic);
    _xic = 0;
  }
  if(_xim)
  {
    XCloseIM(_xim);
    _xim = 0;
  }
  //::XCloseDisplay(_display);
  //_display = 0;
}


typedef struct NativeWndData {
  int width, height;
//  int btnDown;
  int fullScreen: 1;
//  int trackMouse: 1;
  int visible: 1;
  struct {
    float l, t, r, b;
  } dirty;

} NativeWndData;

static NativeWndData* getWndData(NativeWnd w) {
  XPointer ptr = 0;
  if(0 == XFindContext(_display, w, _wdContext, &ptr)) {
    return (NativeWndData*)ptr;
  }
  return 0;
}

static int btnNum(unsigned int btn) {
  if (btn == Button1) {
    return WINL_MOUSE_BTN_LEFT;
  } else if(btn == Button3) {
    return WINL_MOUSE_BTN_RIGHT;
  }
  return 0;
}

static Bool grabPointer(Window win) {
	return XGrabPointer(_display,
		win,
		0,
		PointerMotionMask|
		ButtonMotionMask|
		Button1MotionMask|
		Button2MotionMask|
		ButtonPressMask |
		ButtonReleaseMask,
		GrabModeAsync,
		GrabModeAsync,
		None,
		None,
		CurrentTime) == GrabSuccess;
}

static void ungrabPointer() {
	XUngrabPointer(_display, CurrentTime);
}


static void _windowProc(XEvent* _event)
{
  Window win = _event->xany.window;
  switch(_event->type) {
  case EnterNotify: {
    XCrossingEvent *ce = (XCrossingEvent*) _event;
    winl_on_mouse_enter(win, ce->x, ce->y);
  } break; case LeaveNotify: {
    XCrossingEvent *ce = (XCrossingEvent*) _event;
    winl_on_mouse_leave(win, ce->x, ce->y);
  } break; case ConfigureNotify: {
      int w = _event->xconfigure.width;
      int h = _event->xconfigure.height;
      NativeWndData* wd = getWndData(win);
      // ConfigureNotify is not only for resize purpose, we check if size really changed.
      if(w != wd->width || h != wd->height) {
        wd->width = w; wd->height = h;
        winl_on_resize(win, w, h);
      }
  } break; case MapNotify: {
    NativeWndData* wd = getWndData(win);
    wd->visible = 1;
    float w, h;
    winl_get_size(win, &w, &h);
    winl_on_resize(win, w, h);
  } break; case UnmapNotify: {
    NativeWndData* wd = getWndData(win);
    wd->visible = 0;
    //OnVisibleChanged(False);
  } break; case Expose: {
    winl_expose(win, (float)_event->xexpose.x, (float)_event->xexpose.y,
     (float)_event->xexpose.width, (float)_event->xexpose.height);
     // send later in event loop
  } break; case ButtonPress: {
    XButtonEvent *be = (XButtonEvent*) _event;
    int btn = btnNum(be->button);
    if (btn) {
      winl_on_mouse_press(win, btn, be->x, be->y);
    }
  } break; case ButtonRelease: {
    XButtonEvent *be = (XButtonEvent*) _event;
    int btn = btnNum(be->button);
    if (btn) {
  		winl_on_mouse_release(win, btn, be->x, be->y);
    }
  } break; case MotionNotify:{
    XMotionEvent* me = (XMotionEvent*) _event;
    winl_on_mouse_move(win, me->x, me->y);
  } break; case KeyPress: {
//     {
//       KeySym keysym = 0;
//       Bool charsValid = False;
//       Bool keysymValid = False;
//       String text;
//       if(_xic)
//       {
//         int chars = 0;
//         int buf_size = 64;
//         wchar_t * buf = new wchar_t[buf_size + 1];
//         buf[0] = 0;
//         Bool again = True;
//         while(again)
//         {
//           buf[0] = 0;
//           keysym = 0;
//           Status status;
//           chars = XwcLookupString(_xic, (XKeyPressedEvent *)&_event->xkey, buf, buf_size, &keysym, &status);
//           //hx_trace("[e_ui] XwcLookupString()");
//           switch(status)
//           {
//           case XBufferOverflow:
//             //hx_trace("[e_ui]     XBufferOverflow");
//             buf_size+= 64;
//             delete[] buf;
//             buf = new wchar_t[buf_size + 1];
//             continue;
//           case XLookupNone:
//             //hx_trace("[e_ui]     XLookupNone");
//             again = False;
//             break;
//           case XLookupChars:
//             //hx_trace("[e_ui]     XLookupChars");
//             charsValid = True;
//             again = False;
//             break;
//           case XLookupKeySym:
//             //hx_trace("[e_ui]     XLookupKeySym");
//             keysymValid = True;
//             again = False;
//             break;
//           case XLookupBoth:
//             //hx_trace("[e_ui]     XLookupBoth");
//             charsValid = True;
//             keysymValid = True;
//             again = False;
//             break;
//           }
//         }
//
//         buf[chars] = 0;
//         text = String(buf);
//         delete[] buf;
//       }
//
//       // xim dosn't work, try x11 method
//       if(_xic == 0 || (!charsValid && !keysymValid))
//       {
//         //hx_trace("[e_ui] XLookupString ");
//         //keysym = XKeycodeToKeysym(_display, _event->xkey.keycode, 0);
//         keysymValid = True;
//         Buffer buf1(64);
//         //buf1.(64);
//         //Status status;
//         int len = XLookupString(&_event->xkey, buf1, 64, &keysym, 0);
//         buf1[len] = 0;
//         text = String(buf1);
//       }
//
//       // process keysym
//       if(keysymValid)
//       {
// //				if(keysym == Keyboard::LeftAlt || keysym == Keyboard::RightAlt)
//   //			{
//   //				_altJustDown = True;
//   //			}
//   //			else
//   //			{
//   //				_altJustDown = False;
//   //			}
//
//   //			if(Keyboard::isAltDown())
//   //			{
//   //				OnAltCommand(keysym);
//   //			}
//   //			else
//   //			{
//   //				this->onKeyDown(keysym);
//   //			}
//         if(keysym >= 'a' && keysym <= 'z')
//         {
//           keysym-= 'a' - 'A';
//         }
//         this->onKeyDown(keysym);
//       }
//
//       // process chars
//       if(charsValid)
//       {
//         for(int i = 0; i < text.length(); i++)
//         {
//           this->onCharInput(text[i]);
//         }
//       }
//     }
  } break; case KeyRelease:{
    //{
    //   KeySym keysym = XKeycodeToKeysym(_display, _event->xkey.keycode, 0);
    //
    //   if(Keyboard::isAltDown())
    //   {
    //     // we dont need sys_key_up
    //   }
    //   else
    //   {
    //     if(keysym >= 'a' && keysym <= 'z')
    //     {
    //       keysym-= 'a' - 'A';
    //     }
    //     this->onKeyUp(keysym);
    //   }
    // }
  } break; case FocusIn: {
  	// if(!win->focus)
  	// {
  	// 	win->focus = True;
  	// 	win->OnFocusIn();
  	// }
  } break; case FocusOut:{
  	// if(win->focus){
  	// 	win->focus = False;
  	// 	win->OnFocusOut();
  	// }
  } break; case DestroyNotify: {
    winl_on_destroy(win);
    free(getWndData(win));
    XDeleteContext(_display, win, _wdContext);
    _windowCount--;
    //winl_make_current(0);
  } break; case SelectionRequest: {
    //Clipboard::Singleton().this->HandleSelectionRequest(_event);
  } break; case SelectionNotify: {
  //	Clipboard::Singleton().this->HandleSelectionNotify(_event);
  } break; default: {

  }}
}

static void _HandleEvent(XEvent *_e)
{
	// HX_PROFILE_INCLUDE;
	if(XFilterEvent(_e, None)) {
		return;
	}

	if(_e->type == ClientMessage) {
		XClientMessageEvent* e1 = (XClientMessageEvent*)_e;
		if(e1->message_type == _atom_WM_PROTOCOLS &&
       e1->format == 32 &&
      (Atom)(e1->data.l[0]) == _atom_WM_DELETE_WINDOW) {
      // close button
      XDestroyWindow(_display, e1->window);
		}
    /*
		else if((Atom)(e1.data.l[0]) == _atom_NET_WM_PING)
		{
			WindowImp * imp = _FindWindow(e1.window);
			if(imp != NULL)
			{
				XEvent e2;
				e2.xclient.type         = ClientMessage;
				e2.xclient.display      = _display;
				e2.xclient.message_type = _atom_WM_PROTOCOLS;
				e2.xclient.format       = 32;
				e2.xclient.window       = XDefaultRootWindow(_display);
				e2.xclient.data.l[0]    = e1.data.l[0];
				e2.xclient.data.l[1]    = e1.data.l[1];
				e2.xclient.data.l[2]    = e1.data.l[2];
				e2.xclient.data.l[3]    = 0;
				e2.xclient.data.l[4]    = 0;
				XSendEvent(_display, e2.xclient.window, False, SubstructureRedirectMask|SubstructureNotifyMask, &e2);
			}
		}
    */
		else {
			_windowProc(_e);
		}
	}
	else
	{
		_windowProc(_e);
	}
}

static Bool pumpMessage() {
  XEvent evt;
  if(XEventsQueued(_display, QueuedAfterFlush ) > 0) {
    XNextEvent(_display, &evt);
    _HandleEvent(&evt);
    if (evt.xany.window != 0) {
      NativeWndData *wd = getWndData(evt.xany.window);
      if (wd != 0 && (wd->dirty.l != wd->dirty.r || wd->dirty.t != wd->dirty.b)) {
        winl_on_expose(evt.xany.window, wd->dirty.l, wd->dirty.t, wd->dirty.r-wd->dirty.l, wd->dirty.b-wd->dirty.t);
        wd->dirty.r = wd->dirty.l = wd->dirty.b = wd->dirty.t = 0;
      }
    }
    return True;
  } else {
    return False;
  }
}

int winl_event_loop() {

  winl_on_start();

  while(!_toExit && _windowCount > 1) {
  	if(!pumpMessage()) {
      usleep(1000);
  	}
  }

  _CloseXLib();

	return _exitCode;
}

Bool createGraphics(Window win) {
  if(hGL != NULL)
    return True;

    int visAttributes[20];
    int i = 0;
    visAttributes[i++] = GLX_USE_GL;
    visAttributes[i++] = GLX_USE_GL;
    visAttributes[i++] = GLX_RGBA;
    visAttributes[i++] = GLX_DOUBLEBUFFER;
    visAttributes[i++] = GLX_RED_SIZE;
    visAttributes[i++] = 8;
    visAttributes[i++] = GLX_GREEN_SIZE;
    visAttributes[i++] = 8;
    visAttributes[i++] = GLX_BLUE_SIZE;
    visAttributes[i++] = 8;
    visAttributes[i++] = GLX_ALPHA_SIZE;
    visAttributes[i++] = 8;
    if(False)
    {
      visAttributes[i++] = GLX_STENCIL_SIZE;
      visAttributes[i++] = 8;
    }
    visAttributes[i++] = None;

    //hx_trace("++++++++++++++++++++++");
    XVisualInfo *visinfo = glXChooseVisual(_display, 0, visAttributes);
    // assert(visinfo);
    if(!visinfo)
    {
      return False;
    }
    //hx_trace("----------------------");
    hGL = glXCreateContext(_display, visinfo, NULL, True);
    // assert(hGL);
    if(!hGL)
    {
      return False;
    }

  return True;
}

void winl_get_screen_size(int *width, int *height) {
  if(width) {
    *width = XDisplayWidth(_display, _screenNum);
  }
  if(height) {
    *height = XDisplayHeight(_display, _screenNum);
  }
}

NativeWnd winl_create(int ws, int width, int height) {
  _InitXLib();
  unsigned long valueMask = 0;
  XSetWindowAttributes attr;
  valueMask = CWBorderPixel|CWColormap|CWBitGravity;
  attr.border_pixel = 0;
  attr.colormap = _screenColormap;
  attr.bit_gravity = 0;

  int sw, sh;
  winl_get_screen_size(&sw, &sh);
  if (width <= 0) {
    width = sw / 2;
  }
  if (height <= 0) {
    height = sh / 2;
  }

  // put at center if Window Manager allow
  int x = (sw - width) / 2;
  int y = (sh - height) / 2;

  Window win = XCreateWindow(
    _display,
    RootWindow(_display, _screenNum),
    x, y, width, height,
    0,
    0,
    InputOutput,
    0,
    valueMask,
    &attr
    );

  if(win == 0)
    return 0;

  NativeWndData* wd = malloc(sizeof(NativeWndData));
  memset(wd, 0, sizeof(NativeWndData));
  XSaveContext(_display, win, _wdContext, (XPointer)wd);

  _windowCount++;
  //XClassHint hint;
  //hint.res_name = (char*)"E_APPLICATION";
  //hint.res_class = (char*)"E_WIN";
  //XSetClassHint(_display, imp->hWnd, &hint);

  XSelectInput(_display, win,
      ExposureMask
      | PointerMotionMask
      // FocusChangeMask
      | ButtonMotionMask
      | Button1MotionMask
      // Button2MotionMask
      | KeyPressMask
      | KeyReleaseMask
      | ButtonPressMask
      | ButtonReleaseMask
      | EnterWindowMask
      | LeaveWindowMask
      | StructureNotifyMask
      // im_event_mask
      );

  // add delete button
  XSetWMProtocols (_display, win, &_atom_WM_DELETE_WINDOW, 1);

  if( (ws & WINL_HINT_RESIZABLE) == 0 ) {
    XSizeHints size_hints;
    size_hints.flags = PMinSize | PMaxSize;
    size_hints.min_width = width;
    size_hints.max_width = width;
    size_hints.min_height = height;
    size_hints.max_height = height;
    XSetWMNormalHints(_display, win, &size_hints);
  }

  XSetICValues(_xic, XNClientWindow, win, NULL);
//	WindowImp::_xicWindow = imp->hWnd;
  XSetICFocus(_xic);

  //if(_CreateGraphics())
  //{
  //	return True;
  //}
  //else
  //{
  //	assert(0);
  //	g_window_map.erase(imp->hWnd);
  //	::XDestroyWindow(_display, imp->hWnd);
  //	imp->hWnd = 0;
  //	return False;
  //}

  createGraphics(win);

  return win;
}


void winl_show(NativeWnd win) {
  if(win) {
    XMapWindow(_display, win);
  }
}

void winl_destroy(NativeWnd win) {
	if(win) {
    XDestroyWindow(_display, win);
  }
}

int winl_make_current(NativeWnd win) {
  winl_printf("winl_make_current %ld", win);
  if (win) {
    return glXMakeCurrent(_display, win, hGL);
  } else {
    return glXMakeCurrent(_display, None, NULL);
  }
  //return 0;
}

void winl_swap_buffers(NativeWnd win) {
  if (win) {
    glXSwapBuffers(_display, win);
  }
}

void winl_get_size(NativeWnd win, float *width, float *height) {
  int _w, _h;
  if (win) {
    XWindowAttributes attr;
    XGetWindowAttributes(_display, win, &attr);
    _w = attr.width;
    _h = attr.height;
  } else {
    _w = 0; _h = 0;
  }
  if(width) {
    *width = _w;
  }
  if(height) {
    *height = _h;
  }
}

int winl_is_full_screen(NativeWnd win) {
  if(!win) {
    return 0;
  }
  return getWndData(win)->fullScreen;
}

int winl_is_visible(NativeWnd win) {
  if(!win) {
    return 0;
  }
  return getWndData(win)->visible;
}

void winl_toggle_full_screen(NativeWnd win){
  if(!win) {
    return;
  }

  NativeWndData* wd = getWndData(win);
  if (wd->fullScreen) {
    wd->fullScreen = 0;
    XDeleteProperty(_display, win, _atom_NET_WM_STATE);
    return;
  }
  wd->fullScreen = 1;

  Atom arr[2] = {
    _atom_NET_WM_STATE_ABOVE,
    _atom_NET_WM_STATE_FULLSCREEN,
  };

  XChangeProperty(_display, win, _atom_NET_WM_STATE, XA_ATOM, 32,
    PropModeReplace, (const unsigned char *)arr, 2);
}

void winl_set_title(NativeWnd win, const char * title) {
  if(!win) {
    return;
  }
  XTextProperty windowName;
  Status status = Xutf8TextListToTextProperty(_display,
    (char **) &title, 1, XUTF8StringStyle, &windowName);
  if (status == Success) {
      XSetTextProperty(_display, win, &windowName, _atom_NET_WM_NAME);
      XFree(windowName.value);
  }
}

void winl_expose(NativeWnd win, float x, float y, float width, float height) {
  if (!win) {
    return;
  }
  NativeWndData * wd = getWndData(win);
  wd->dirty.l = MIN(x, wd->dirty.l);
  wd->dirty.t = MIN(y, wd->dirty.t);
  wd->dirty.r = MAX(x+width, wd->dirty.r);
  wd->dirty.b = MAX(y+height, wd->dirty.b);
}

void winl_exit_loop(int code) {
  _toExit = 1;
  _exitCode = code;
}

void winl_message_box(NativeWnd win, const char* msg, const char* title) {
	// stub for linux
  winl_printf("%s\n", "stub: winl_message_box");
}

int winl_confirm_box(NativeWnd win, const char* msg, const char* title) {
  // stub for linux
  winl_printf("%s\n", "stub: winl_confirm_box");
}

char* winl_os_version() {
  char* buf = (char*)malloc(4096);
  struct utsname x;
  if (0 == uname(&x)) {
    sprintf(buf, "%s %s %s", x.sysname, x.release, x.version);
  } else {
    buf[0] = 0;
    strcat(buf, "Linux unkown");
  }
  return buf;
}

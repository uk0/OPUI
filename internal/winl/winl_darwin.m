#import <Cocoa/Cocoa.h>
#import "winl-c.h"

@class WindowController;
@class OpenGLView;

/*
██ ███    ██ ████████ ███████ ██████  ███████  █████   ██████ ███████
██ ████   ██    ██    ██      ██   ██ ██      ██   ██ ██      ██
██ ██ ██  ██    ██    █████   ██████  █████   ███████ ██      █████
██ ██  ██ ██    ██    ██      ██   ██ ██      ██   ██ ██      ██
██ ██   ████    ██    ███████ ██   ██ ██      ██   ██  ██████ ███████
*/

@interface AppDelegate : NSObject <NSApplicationDelegate>
@end

static AppDelegate* getAppDelegate() {
  return NSApplication.sharedApplication.delegate;
}

@interface OpenGLView : NSOpenGLView
{
  @public
  WindowController* _wc;
//  NSTrackingArea* _ta;
}
@end

@interface ViewController : NSViewController

@end

// @interface Window : NSWindow
// {
//
// }
// @end

@interface WindowController : NSWindowController<NSWindowDelegate>
{
@public
  OpenGLView* glview;
  BOOL _bFirstResize;
}
@end

/*
██ ███    ███ ██████  ██      ███████ ███    ███ ███████ ███    ██ ████████
██ ████  ████ ██   ██ ██      ██      ████  ████ ██      ████   ██    ██
██ ██ ████ ██ ██████  ██      █████   ██ ████ ██ █████   ██ ██  ██    ██
██ ██  ██  ██ ██      ██      ██      ██  ██  ██ ██      ██  ██ ██    ██
██ ██      ██ ██      ███████ ███████ ██      ██ ███████ ██   ████    ██
*/

@implementation AppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {
  winl_on_start();
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
  return TRUE;
}

- (void)applicationWillTerminate:(NSNotification *)aNotification {
  winl_on_exit(0);
}
@end // AppDelegate


@implementation OpenGLView

- (instancetype)initWithFrame:(NSRect)frameRect pixelFormat:(NSOpenGLPixelFormat *)format {
    self = [super initWithFrame:frameRect pixelFormat:format];
    if (self) {
      NSTrackingArea * ta = [[NSTrackingArea alloc]
        initWithRect:(NSRect)self.bounds
        options: (/*NSTrackingActiveInKeyWindow |*/ NSTrackingActiveAlways | NSTrackingMouseEnteredAndExited |
          NSTrackingMouseMoved | NSTrackingInVisibleRect)
        owner: self
        userInfo: nil];
      [self addTrackingArea: ta];
      [ta release];
    }
    return self;
}
//
// - (oneway void)release {
//   if(self->_ta) {
//     // [self->_ta release];
//   }
//   [super release];
// }

- (BOOL) acceptsFirstResponder {
  return YES;
}

- (BOOL)isOpaque {
    return YES;
}

- (void)drawRect:(NSRect)dirtyRect {
    [super drawRect:dirtyRect];
    float x = dirtyRect.origin.x;
    float y = dirtyRect.origin.y;
    float w = dirtyRect.size.width;
    float h = dirtyRect.size.height;
    y = self.bounds.size.height - y;
    winl_on_expose(self->_wc, x, y, w, h);
}

// - (void)viewDidEndLiveResize {
//   [super viewDidEndLiveResize];
//   CGSize sz = _bounds.size;
//   winl_on_resize(_wc, sz.width, sz.height);
// }

- (void)mouseEntered:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_enter(self->_wc, pt.x, pt.y);
}

- (void)mouseMoved:(NSEvent *)theEvent {
  [super mouseMoved: theEvent];
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_move(self->_wc, pt.x, pt.y);
}

- (void)mouseExited:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_leave(self->_wc, pt.x, pt.y);
}

- (void)mouseDown:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_press(self->_wc, WINL_MOUSE_BTN_LEFT, pt.x, pt.y);
}

- (void)mouseUp:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_release(self->_wc, WINL_MOUSE_BTN_LEFT, pt.x, pt.y);
}

- (void)rightMouseDown:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_press(self->_wc, WINL_MOUSE_BTN_RIGHT, pt.x, pt.y);
}

- (void)rightMouseUp:(NSEvent *)theEvent {
  NSPoint pt = [self convertPoint:[theEvent locationInWindow] fromView:nil];
  pt.y = self.bounds.size.height - pt.y;
  winl_on_mouse_release(self->_wc, WINL_MOUSE_BTN_RIGHT, pt.x, pt.y);
}

@end // OpenGLView

@implementation ViewController

- (void)viewDidLoad {
  [super viewDidLoad];

  // Do any additional setup after loading the view.
}


- (void)setRepresentedObject:(id)representedObject {
  [super setRepresentedObject:representedObject];

  // Update the view, if already loaded.
}

@end // ViewController

// @implementation Window
// - (void)mouseMoved:(NSEvent *)theEvent {
//   NSPoint pt = [theEvent locationInWindow];
//   //winl_on_mouse_move(self, pt.x, pt.y);
//   printf("%s\n",  "mouseMoved");
// }
// @end // Window

@implementation WindowController

- (void)windowDidLoad {
    [super windowDidLoad];
    printf("windowDidLoad\n");
}

- (void)windowWillClose:(NSNotification *)notification {
  winl_on_destroy(self);
}

- (void)windowDidBecomeKey:(NSNotification *)notification {
    [self.window makeFirstResponder: self.window.contentView];
    self.window.acceptsMouseMovedEvents = YES;
    //printf("%s\n", "windowDidBecomeKey");
}

- (void)windowDidChangeOcclusionState:(NSNotification *)notification {
  if (!self->_bFirstResize){
    CGSize sz = glview.frame.size;
    winl_on_resize(self, sz.width, sz.height);
    self->_bFirstResize = 1;
  }
}

- (void)windowDidResize:(NSNotification *)notification {
  self->_bFirstResize = 1;
  CGSize sz = glview.frame.size;
  winl_on_resize(self, sz.width, sz.height);
}

- (void)windowDidMiniaturize:(NSNotification *)notification {

}

- (void)windowDidDeminiaturize:(NSNotification *)notification {

}

- (void)windowDidEnterFullScreen:(NSNotification *)notification {

}

- (void)windowDidExitFullScreen:(NSNotification *)notification {

}
@end // WindowController


/*
 ██████     ███████ ██    ██ ███    ██  ██████ ████████ ██  ██████  ███    ██ ███████
██          ██      ██    ██ ████   ██ ██         ██    ██ ██    ██ ████   ██ ██
██          █████   ██    ██ ██ ██  ██ ██         ██    ██ ██    ██ ██ ██  ██ ███████
██          ██      ██    ██ ██  ██ ██ ██         ██    ██ ██    ██ ██  ██ ██      ██
 ██████     ██       ██████  ██   ████  ██████    ██    ██  ██████  ██   ████ ███████
*/

int winl_event_loop() {
  NSApplication * app = [NSApplication sharedApplication];
  // set the app delegate, don't need release.
  [app setDelegate:[[AppDelegate alloc] init]];
  // make window popup in the front as normal app
  [app setActivationPolicy:NSApplicationActivationPolicyRegular];
  [app activateIgnoringOtherApps: YES];
  [app run];
  return EXIT_SUCCESS;
}

void winl_exit_loop(int code) {
  [NSApplication.sharedApplication terminate:nil];
}

NSOpenGLContext* findOpenGLContex() {
  NSArray<NSWindow*>* windows = NSApplication.sharedApplication.windows;
  OpenGLView *glview = nil;
  for(NSUInteger i=0; i < windows.count; i++) {
    id view = windows[i].contentView;
    if ([view isKindOfClass:[OpenGLView class]]) {
      glview = view;
      break;
    }
  }
  if (glview == nil) {
    return nil;
  }
  return glview.openGLContext;
}

NativeWnd winl_create(int ws, int width, int height) { @autoreleasepool{


  NSOpenGLContext* sharedOpenGLContex = findOpenGLContex();

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

  NSRect frame = NSMakeRect(x, y, width, height);
  NSWindowStyleMask mask = NSWindowStyleMaskClosable | NSWindowStyleMaskTitled | NSWindowStyleMaskMiniaturizable;
  if(ws & WINL_HINT_RESIZABLE)
    mask |= NSWindowStyleMaskResizable;
  NSWindow* window = [[NSWindow alloc] initWithContentRect:frame
                                                  styleMask: mask
                                                    backing:NSBackingStoreBuffered
                                                      defer:NO];

//  AppDelegate* delegate = getAppDelegate();
  //NSOpenGLPixelFormat* pf = (ws & WINL_HINT_ANIMATE) ? delegate->_animatePixelFormat : delegate->_staticPixelFormat;
  NSOpenGLPixelFormatAttribute animateAttributes[] = {
    NSOpenGLPFAOpenGLProfile, NSOpenGLProfileVersionLegacy,
    NSOpenGLPFAColorSize, (NSOpenGLPixelFormatAttribute) 24,
    NSOpenGLPFAAlphaSize, (NSOpenGLPixelFormatAttribute) 8,
    NSOpenGLPFADepthSize, (NSOpenGLPixelFormatAttribute) 0,
    NSOpenGLPFAStencilSize, 0,
    NSOpenGLPFADoubleBuffer,
    NSOpenGLPFAAccelerated,
    0
  };
  NSOpenGLPixelFormat* pf = [[NSOpenGLPixelFormat alloc] initWithAttributes: animateAttributes];

  OpenGLView* view = [[OpenGLView alloc] initWithFrame: [window frame] pixelFormat: pf];
  // [pf release];
  if (sharedOpenGLContex != nil) {
    NSOpenGLContext *ctx = [[NSOpenGLContext alloc]
                    initWithFormat: pf
                    shareContext:sharedOpenGLContex];
    view.openGLContext = ctx;
  } else {
    sharedOpenGLContex = view.openGLContext;
  }

  ViewController* vc = [[ViewController alloc] initWithNibName: nil bundle: nil];
  vc.view = view;
  window.contentViewController = vc;
  WindowController* wc = [[WindowController alloc] initWithWindow: window];
  window.delegate = wc;

  //wc->_assoc_data = assoc_data;
  view->_wc = wc;
  wc->glview = view;

  // NSUInteger count = [[[NSApplication sharedApplication] windows] count];
  return (NativeWnd)wc;
}}

void winl_get_screen_size(int *width, int *height) {
  NSScreen *mainScreen = [NSScreen mainScreen];
  NSRect screenRect = [mainScreen visibleFrame];
  if (width) {
    *width = screenRect.size.width;
  }
  if (height) {
    *height = screenRect.size.height;
  }
}

void winl_show(NativeWnd win) {
  //AppDelegate* delegate = getAppDelegate();
  WindowController* wc = (WindowController*)win;
  if(wc) {
      [wc.window makeKeyAndOrderFront: nil];
      //[wc->glview viewDidEndLiveResize];
  }
}

void winl_destroy(NativeWnd win) {
  //AppDelegate* delegate = getAppDelegate();
  WindowController* wc = (WindowController*)win;
  if(wc) {
    [wc close];
  }
}

int winl_make_current(NativeWnd win) {
  //AppDelegate* delegate = getAppDelegate();
  WindowController* wc = (WindowController*)win;
  if(wc) {
    [wc->glview.openGLContext makeCurrentContext];
  }
  return 1;
}

void winl_swap_buffers() {
  [[NSOpenGLContext currentContext] flushBuffer];
}


int winl_is_full_screen(NativeWnd win) {
  WindowController* wc = (WindowController*)win;
  if (wc) {
    return (wc.window.styleMask & NSWindowStyleMaskFullScreen) == NSWindowStyleMaskFullScreen;
  }
  return 0;
}

int winl_is_visible(NativeWnd win) {
  WindowController* wc = (WindowController*)win;
  if (wc) {
    return wc.window.visible;
  }
  return 0;
}

void winl_toggle_full_screen(NativeWnd win) {
    WindowController* wc = (WindowController*)win;
    if (wc) {
      [wc.window toggleFullScreen: nil];
    }
}

// void winl_get_mouse_pos(NativeWnd win, float *x, float *y) {
//
// }

void winl_set_title(NativeWnd win, const char * title) {
    WindowController* wc = (WindowController*)win;
    if (wc) {
      NSString* s = [[NSString alloc] initWithUTF8String: title];
      wc.window.title = s;
      [s release];
    }
}

void winl_expose(NativeWnd win, float x, float y, float width, float height) {
    WindowController* wc = (WindowController*)win;
    if (!wc) {
      return;
    }
    y =  wc->glview.bounds.size.height - y;
    NSRect invalidRect = NSMakeRect(x, y, width, height);
    [wc->glview setNeedsDisplayInRect: invalidRect];
}


static int messasge_box(WindowController* wc, const char* msg, const char* title, int confirm) {
	NSString* ms = [[NSString alloc] initWithUTF8String: msg];
  NSString* ts = [[NSString alloc] initWithUTF8String: title];

  NSAlert *alert = [[NSAlert alloc] init];
  // TODO: translate?
  [alert addButtonWithTitle:@"OK"];
  if (confirm) {
    [alert addButtonWithTitle:@"Cancel"];
    [alert setAlertStyle:NSAlertStyleWarning];
    // TODO: icon
  }

  [alert setMessageText: ts];
  [alert setInformativeText: ms];

  NSModalResponse res = [alert runModal];

  [alert release];

	[ms release];
	[ts release];

	return res == NSAlertFirstButtonReturn;
}

void winl_message_box(NativeWnd win, const char* msg, const char* title) {
	messasge_box((WindowController*)win, msg, title, 0);
}

int winl_confirm_box(NativeWnd win, const char* msg, const char* title) {
	if(1 == messasge_box((WindowController*)win, msg, title, 1)) {
		return 1;
	} else {
		return 0;
	}
}

char* winl_os_version() {
  NSOperatingSystemVersion version = NSProcessInfo.processInfo.operatingSystemVersion;
  char* buf = (char*)malloc(128);
  sprintf(buf, "MacOS %ld.%ld.%ld", version.majorVersion, version.minorVersion, version.patchVersion);
  return buf;
}

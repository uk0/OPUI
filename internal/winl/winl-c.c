
#include <stdarg.h>
#include <stdio.h>
#include "winl-c.h"

int winl_printf(const char *fmt, ...) {
  char buf[4096];
  va_list ap;
  int ret;
  va_start(ap, fmt);
  ret = vsprintf(buf, fmt, ap);
  va_end(ap);
  winl_report(buf, 0);
  return ret;
}

int winl_panicf(const char *fmt, ...) {
  char buf[4096];
  va_list ap;
  int ret;
  va_start(ap, fmt);
  ret = vsprintf(buf, fmt, ap);
  va_end(ap);
  winl_report(buf, 1);
  return ret;
}

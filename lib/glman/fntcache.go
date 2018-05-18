package glman

import "time"

const (
	durFreeFont = time.Second
)

var fntCache = make(map[Font]*fcItem)
var timeFntMantain time.Time

type fcItem struct {
	f     *texFont
	atime time.Time
}

func accessFont(name Font) (f *texFont) {
	now := time.Now()
	defer func() {
		if now.Before(timeFntMantain) || now.Sub(timeFntMantain) > durFreeFont {
			timeFntMantain = now
			var del []Font
			for k, item := range fntCache {
				if item.f != f && (now.Before(item.atime) || now.Sub(item.atime) > durFreeFont) {
					del = append(del, k)
					item.f.finalize()
				}
			}
			for _, k := range del {
				delete(fntCache, k)
			}
		}
	}()
	p, ok := fntCache[name]
	if ok {
		p.atime = now
		f = p.f
		return
	}

	f = new(texFont)
	f.init(name)
	fntCache[name] = &fcItem{f: f, atime: now}
	return f
}

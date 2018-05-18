package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"tetra/internal/winl"
	"tetra/lib/lang"
	"tetra/lib/store"
)

//go:generate go run ../../cmd/lang-extract/lang-extract.go ../../testdata/dict ..

var (
	overLc string
	force  bool

	config Config
)

func init() {
	log.SetFlags(log.Ltime)

	flag.StringVar(&overLc, "locale", "",
		`Locale code like "en-US", leave it empty to auto detect.`)
	flag.BoolVar(&force, "force", false,
		`Don not prompt confirm dialogs, imply "Yes" or "OK".`)
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", name)
	//fmt.Fprintf(os.Stderr, "  %s [OPTIONS] {DST_DIR} {SRC_DIR}\n\n", name)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
}

// Run run the main event loop.
func Run(onStart func() error, onExit func(err error)) {
	// parse the command line options
	flag.Usage = usage
	flag.Parse()

	for _, d := range []string{"log", "dict", "state", "layout"} {
		store.MkdirAll(d)
	}

	store.LoadState("", "app", &config)
	if overLc != "" {
		config.Locale = overLc
	}
	config.Locale, _ = lang.Load("dict", config.Locale)

	log.Println("OS version:", winl.OSVersion())

	winl.Run(func() (err error) {
		if onStart != nil {
			err = onStart()
		}
		return err
	}, func(err error) {
		if onExit != nil {
			onExit(err)
		}
		store.SaveState("", "app", &config)
	})

	// we should not put code after the Run func, because program had likely triminated
}

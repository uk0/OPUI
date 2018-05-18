package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"tetra/lib/lang"
)

var (
	prune      bool
	verbose    bool
	localeList string
	obsolete   string
	keywords   string
	filetypes  string

	rxtemplate = `\(\s*((?s:\x60[^\x60]*\x60)|"(?:(?:\\"|.)*?)")\s*\)`
	//   keyword   (        ` text no bq   `                        )
	//   keyword   (                          " text \" quote "     )
	regexpTr *regexp.Regexp

	ftset map[string]bool
)

func init() {
	flag.BoolVar(&prune, "prune", false, "Delete obsolete items.")
	flag.BoolVar(&verbose, "v", false, "Print verbose messages.")
	flag.StringVar(&localeList, "locales", `en-US zh-CN`, "Space separated locale list to generate.")
	flag.StringVar(&obsolete, "obsolete", `~^(del)`, "Surffix for obsolete items.")
	flag.StringVar(&keywords, "keys", "lang . Tr |lang . Or |Gettext ",
		"'|' separated keywords to extract, space is interpret as any number of spaces.")
	flag.StringVar(&filetypes, "types", "go", "Space separated file extensons to treat as sources.")
}

func isSources(name string) bool {
	ext := filepath.Ext(name)
	_, ok := ftset[ext]
	return ok
}
func recursiveExtract(words map[string]string, dir string) {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, v := range entries {
		name := v.Name()
		if name[0] == '.' {
			continue
		}
		if v.IsDir() {
			recursiveExtract(words, dir+"/"+name)
		} else if isSources(name) {
			extract(words, dir+"/"+name)
		}
	}
}

func extract(words map[string]string, filename string) {
	// note: we can not use go/parser package because we must support ill form codes
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	var pathPrinted bool
	all := regexpTr.FindAllStringSubmatch(string(s), -1)
	for _, sm := range all {
		word := sm[1]
		if word != "" && len(word) > 2 {
			word, err = strconv.Unquote(word)
			if !pathPrinted && (err != nil || verbose) {
				fmt.Printf("  %s\n", filename)
				pathPrinted = true
			}
			if err != nil {
				fmt.Printf("    ! %s\n", strconv.Quote(sm[1]))
				continue
			}
			if verbose {
				fmt.Printf("    + %s\n", strconv.Quote(word))
			}
			words[word] = filename
		}
	}
}

func addObsolete(s string) string {
	if !strings.HasSuffix(s, obsolete) {
		return s + obsolete
	}
	return s
}

func removeObsolete(s string) string {
	return strings.TrimSuffix(s, obsolete)
}

func outputFile(dir, lc string, words map[string]string) {
	if verbose {
		fmt.Printf("  %s.json\n", lc)
	}
	var old map[string]string
	data, err := ioutil.ReadFile(dir + "/" + lc + ".json")
	if err == nil {
		json.Unmarshal(data, &old)
	}

	var dict = make(map[string]string)
	for w := range words {
		dict[w] = ""
	}
	for w, v := range old {
		w = removeObsolete(w)
		if _, ok := dict[w]; !ok {
			if prune {
				continue
			}
			dict[addObsolete(w)] = v
			continue
		}
		dict[w] = v
	}
	buf, err := json.MarshalIndent(dict, "", "  ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(dir+"/"+lc+".json", buf, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write dst dir: %v", err)
		os.Exit(1)
	}
}

func outputDir(dir string, words map[string]string) {
	locales := make(map[string]bool)

	os.MkdirAll(dir, 0755)

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to read dst dir: %v", err)
		os.Exit(1)
	}
	for _, v := range entries {
		name := v.Name()
		if name[0] == '.' || strings.HasSuffix(name, `.json`) || v.IsDir() {
			continue
		}
		locales[strings.TrimSuffix(name, `.json`)] = true
	}

	for _, lc := range strings.Split(localeList, ` `) {
		lc = strings.TrimSpace(lc)
		if _, ok := locales[lc]; ok {
			continue
		}
		if !lang.IsValid(lc) {
			fmt.Fprintf(os.Stderr, "warning: will not generate for invalid locale \"%s\"\n", lc)
			continue
		}
		locales[lc] = true
	}
	if len(locales) == 0 {
		fmt.Printf("no target locales.\n")
		return
	}
	if verbose {
		fmt.Printf("generating in %s:\n", dir)
	}
	for lc := range locales {
		outputFile(dir, lc, words)
	}
	if verbose {
		fmt.Printf("  %d files generated.\n", len(locales))
	}
}

func genRegexpr() {
	s := keywords
	s = strings.Replace(s, ` `, `\s*`, -1) // spaces
	//s = strings.Replace(s, `,`, `|`, -1)   // regexp or
	s = strings.Replace(s, `.`, `\.`, -1)
	if strings.IndexByte(keywords, ',') != -1 {
		s = `(?:` + s + `)`
	}
	s = `\b` + s + rxtemplate
	if verbose {
		fmt.Printf("regexpr: %s\n", s)
	}
	regexpTr = regexp.MustCompile(s)
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", name)
	fmt.Fprintf(os.Stderr, "  %s [OPTIONS] {DST_DIR} {SRC_DIR}\n\n", name)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	ftset = make(map[string]bool)
	for _, s := range strings.Split(filetypes, " ") {
		if !strings.HasPrefix(s, ".") {
			s = `.` + s
		}
		ftset[s] = true
	}
	genRegexpr()

	dstDir := flag.Arg(0)
	srcDir := flag.Arg(1)

	words := make(map[string]string)
	if verbose {
		fmt.Printf("extracting from %s:\n", srcDir)
	}
	recursiveExtract(words, srcDir)
	if verbose {
		fmt.Printf("  %d words extracted.\n", len(words))
	}

	outputDir(dstDir, words)
}

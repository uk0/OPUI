package lang

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"tetra/lib/dbg"
	"tetra/lib/levenshtein"
	"tetra/lib/store"
)

var (
	dict = make(map[string]string)

	// store words which already printed debug message
	debug = make(map[string]bool)
	rxlc  = regexp.MustCompile(`^[a-z][a-z](?:[\-_][A-Z][A-Z]?)$`)

	// ErrBadLocale indicate locale is invalid
	ErrBadLocale = errors.New(`Bad locale`)

	matchlc func(availables, prefers []string) string
	validlc func(string) bool
)

func filename(dir, lc string) string {
	return dir + `/` + lc + `.json`
}

// Reset to initial state
func Reset() {
	dict = make(map[string]string)
}

// Load dictionary, do fuzzy match if lc unavaible.
func Load(dir, lc string) (matched string, err error) {
	if lc == "" {
		lc = Detect()
		dbg.Logf("use auto detected locale %s\n", lc)
	}
	dbg.Logf("loading dict \"%s\" from %s\n", lc, dir)
	defer func() {
		matched = lc
		if err != nil {
			log.Printf("error: failed to load dict \"%s\" from %s\n", lc, dir)
		}
	}()
	err = loadExact(dir, lc)
	if err == nil {
		return
	}
	lcs := Enum(dir)
	lc = Match(lcs, lc)
	dbg.Logf("  use %s\n", lc)
	err = loadExact(dir, lc)
	return
}

func loadExact(dir, lc string) error {
	if !IsValid(lc) {
		return ErrBadLocale
	}
	var fn = filename(dir, lc)
	var m map[string]string
	var data []byte
	var err error

	data, err = store.ReadFile(fn)

	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	for k, v := range m {
		if v == "" {
			dbg.Logf("%s: untranslated: %s\n", filepath.Base(fn), strconv.Quote(v))
			continue
		}
		dict[k] = v
	}
	return nil
}

// Enum enum available locales
func Enum(dir string) (lcs []string) {
	var ents []os.FileInfo
	var err error
	ents, err = store.ReadDir(dir)
	if err != nil {
		return
	}
	for _, ent := range ents {
		name := ent.Name()
		if ent.IsDir() || !strings.HasSuffix(name, `.json`) {
			continue
		}
		name = strings.TrimSuffix(name, `.json`)
		if !IsValid(name) {
			continue
		}
		lcs = append(lcs, name)
	}
	return
}

// Tr translate str
func Tr(str string) string {
	if str == "" {
		return str
	}
	if dst, ok := dict[str]; ok {
		return dst
	}
	dict[str] = str
	dbg.Logf("untranslated text: %s\n", strconv.Quote(str))
	return str
}

// Or simply return the orintial str
func Or(str string) string {
	return str
}

// IsValid check whether lc is valid lc
func IsValid(lc string) bool {
	if !rxlc.MatchString(lc) {
		return false
	}
	if validlc != nil {
		return validlc(lc)
	}
	return true
}

func parseEnvStr(s string) string {
	if s == "" {
		return ""
	}
	pos := strings.IndexByte(s, '.')
	if pos != -1 {
		s = s[:pos]
	}
	if IsValid(s) {
		return s
	}
	return ""
}

// Detect auto detect locle.
func Detect() (s string) {
	defer func() {
		s = strings.Replace(s, "_", "-", -1)
	}()
	s = parseEnvStr(os.Getenv("LANG"))
	if s != "" {
		return s
	}
	s = parseEnvStr(os.Getenv("LC_ALL"))
	if s != "" {
		return s
	}
	s = "en"
	return
}

// Match find prefer locale from availables
func Match(availables []string, prefers ...string) (lc string) {
	if len(availables) == 0 {
		return "en"
	}
	if len(availables) == 1 {
		return availables[0]
	}
	if len(prefers) == 0 {
		return availables[0]
	}

	if matchlc != nil {
		lc = matchlc(availables, prefers)
		if lc != "" {
			return
		}
	}

	d := int(math.MaxInt32)
	lc = availables[0] // fallback
	for i, p := range prefers {
		p = strings.ToLower(p)
		for _, a := range availables {
			a1 := strings.ToLower(a)
			d2 := len(a1) - len(a1)
			if d2 < 0 {
				d2 = -d2
			}
			d1 := (i + 1) * (levenshtein.Distance(p, a1)*3 + d2)
			if d1 < d {
				lc = a
				d = d1
			}
		}
	}
	return lc
}

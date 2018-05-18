package linereader

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type srcError struct {
	file    string
	line    int
	snippet string
	msg     string
}

func (e *srcError) Error() string {
	return fmt.Sprintf("%s:%d: error: %s\nin:\n%s", e.file, e.line, e.msg, e.snippet)
}

// LineReader read text stream line by line
type LineReader struct {
	R        *bufio.Reader
	FileName string
	LineNum  int
}

// ReadLine read next line, for example lr.ReadLine('\n', true, true)
func (lr *LineReader) ReadLine(lineEnd byte, trim bool, skipEmpty bool) (s string, err error) {
	for {
		lr.LineNum++
		s, err = lr.R.ReadString(lineEnd)
		if err != nil && s == "" {
			return "", err
		}
		err = nil
		if trim {
			s = strings.TrimSpace(s)
		}
		if skipEmpty && s == "" {
			continue
		}
		return
	}
}

// Error create an error with current file name and line number information
func (lr *LineReader) Error(snippet, msg string) error {
	return &srcError{lr.FileName, lr.LineNum, snippet, msg}
}

// New create LineReader from io.Reader,
//   note: use &LineReader{br, filename, 0} to create from bufio.Reader
func New(r io.Reader, filename string) *LineReader {
	br := bufio.NewReader(r)
	return &LineReader{br, filename, 0}
}

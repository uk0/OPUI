package store

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"tetra/lib/dbg"
	"time"
)

var (
	// writable directory
	dat = filepath.Clean(detectDataPath())
	// readonly directory, can be ""
	res = filepath.Clean(detectResPath())

	logfile io.WriteCloser
)

func detectResPath() string {
	// TODO: implement
	return "testdata"
}

func detectDataPath() string {
	// TODO: implement
	return "testdata"
}

// a fork writer for write to console and log file at the same time
type forkWriter struct {
	w1 io.Writer
	w2 io.Writer
}

func (w *forkWriter) Write(b []byte) (int, error) {
	w.w2.Write(b)
	return w.w1.Write(b)
}

func forkLog() error {
	filename := time.Now().Format("log/2006-01-02.log")
	var err error
	logfile, err = Append(filename)
	if err != nil {
		log.SetOutput(os.Stderr)
		dbg.Logf("failed to append to log file %s\n", filename)
		return err
	}
	//logfile.Write(mmlog.Bytes())
	log.SetOutput(&forkWriter{os.Stderr, logfile})
	dbg.Logf("redirect log to json %s\n", filename)
	return nil
}

func ResPath(sub string) string {
	if sub == "" {
		return res
	}
	if strings.HasPrefix(sub, "/") {
		return res + sub
	}
	return res + "/" + sub
}

func DataPath(sub string) string {
	if sub == "" {
		return dat
	}
	if strings.HasPrefix(sub, "/") {
		return dat + sub
	}
	return dat + "/" + sub
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func Stat(name string) (info os.FileInfo, err error) {
	info, err = os.Stat(DataPath(name))
	if err != nil && res != "" {
		info, err = os.Stat(ResPath(name))
	}
	return
}

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
func MkdirAll(dir string) error {
	return os.MkdirAll(DataPath(dir), 0777)
}

// ReadDir reads the directory named by dirname and returns
// a list of directory entries sorted by filename.
func ReadDir(dir string) ([]os.FileInfo, error) {
	v1, e1 := ioutil.ReadDir(DataPath(dir))
	if res == "" {
		return v1, e1
	}

	v2, e2 := ioutil.ReadDir(ResPath(dir))
	if e1 != nil && e2 != nil {
		return nil, e1
	}
	if e1 == nil && e2 != nil {
		return v1, nil
	}
	if e1 != nil && e2 == nil {
		return v2, nil
	}
	set := make(map[string]bool)
	for _, x := range v1 {
		set[x.Name()] = true
	}
	for _, x := range v2 {
		if _, ok := set[x.Name()]; !ok {
			v1 = append(v1, x)
		}
	}
	return v1, nil
}

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile
// reads the whole file, it does not treat an EOF from Read as an error
// to be reported.
func ReadFile(filename string) (b []byte, err error) {
	b, err = ioutil.ReadFile(DataPath(filename))
	if err != nil && res != "" {
		b, err = ioutil.ReadFile(ResPath(filename))
	}
	return
}

// WriteFile writes data to a file named by filename.
// If the file does not exist, WriteFile creates it with permissions perm;
// otherwise WriteFile truncates it before writing.
func WriteFile(filename string, data []byte) error {
	return ioutil.WriteFile(DataPath(filename), data, 0666)
}

// Append open file in append mode, create is not exist
func Append(filename string) (*os.File, error) {
	return os.OpenFile(DataPath(filename), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned
// File can be used for I/O; the associated file descriptor has mode
// O_RDWR.
// If there is an error, it will be of type *PathError.
func Create(filename string) (*os.File, error) {
	return os.Create(DataPath(filename))
}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func Open(filename string) (file *os.File, err error) {
	file, err = os.Open(DataPath(filename))
	if err != nil && res != "" {
		file, err = os.Open(ResPath(filename))
	}
	return
}

// Truncate changes the size of the named file.
// If the file is a symbolic link, it changes the size of the link's target.
// If there is an error, it will be of type *PathError.
func Truncate(name string, size int64) error {
	return os.Truncate(DataPath(name), size)
}

// Remove removes the named file or directory.
// If there is an error, it will be of type *PathError.
func Remove(name string) error {
	return os.Remove(DataPath(name))
}

// SaveState save v to {store}/dir/name.json, use json.MarshalIndent.
func SaveState(dir, name string, v interface{}) (err error) {
	name = dir + "/" + url.QueryEscape(name) + ".json"
	defer func() {
		if err == nil {
			dbg.Logf("succeeded save state: %s\n", name)
		} else {
			dbg.Logf("failed save state: %s, error: %v\n", name, err)
		}
	}()
	var data []byte
	if x, ok := v.(interface {
		State() ([]byte, error)
	}); ok {
		data, err = x.State()
	} else {
		data, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return
		}
	}
	err = WriteFile(name, data)
	return
}

// LoadState load v from {store}/dir/name.json, use json.Unmarshal
func LoadState(dir, name string, v interface{}) (err error) {
	name = dir + "/" + url.QueryEscape(name) + ".json"
	defer func() {
		if err == nil {
			dbg.Logf("succeeded load state: %s\n", name)
		} else {
			dbg.Logf("failed load state: %s, error: %v\n", name, err)
		}
	}()
	data, err := ReadFile(name)
	if err != nil {
		return
	}
	if x, ok := v.(interface {
		SetState([]byte) error
	}); ok {
		err = x.SetState(data)
	} else {
		err = json.Unmarshal(data, v)
	}
	return
}

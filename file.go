package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	os.FileInfo
	config  Config
	dirname string
	parent  *Dir

	selected bool
}

func NewFile(path string, config Config) (*File, error) {
	var err error
	f := &File{
		config:  config,
		dirname: filepath.Dir(path),
	}
	f.FileInfo, err = os.Stat(path)
	if err != nil {
		return nil, err
	}
	if f.FileInfo.IsDir() {
		return nil, fmt.Errorf("the path '%s' isn't file", path)
	}
	return f, err
}

func (f *File) Type() string {
	return "file"
}

func (f *File) SetParent(p *Dir) {
	f.parent = p
}

func (f *File) Parent() *Dir {
	return f.parent
}

func (f *File) Path() string {
	return filepath.Join(f.dirname, f.Name())
}

func (f *File) Selected() bool {
	return f.selected
}

func (f *File) Select() {
	f.selected = true
}

func (f *File) Unselect() {
	f.selected = false
}

func (f *File) ToggleSelected() {
	f.selected = !f.selected
}

func (f *File) Rename(newName string) error {
	return os.Rename(f.Path(), filepath.Join(f.dirname, newName))
}

func (f *File) Move(newDirname string) error {
	return os.Rename(f.Path(), filepath.Join(newDirname, f.Name()))
}

func (f *File) Remove() error {
	return os.Remove(f.Path())
}

// Non interface methods

func (f *File) Equals(t *File) bool {
	if t == nil {
		return false
	}
	return f.Path() == t.Path()
}

func (f *File) line(depth int) []byte {
	var prefix string
	if f.selected {
		prefix = f.config.PrefixSelected
	} else {
		prefix = f.config.PrefixFile
	}
	return []byte(fmt.Sprintf("%s%s %s", strings.Repeat(f.config.Indent, depth), prefix, f.Name()))
}

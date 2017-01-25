package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	context *Context

	os.FileInfo
	dirname string
	parent  *Dir

	selected bool
}

func NewFile(path string, context *Context) (*File, error) {
	var err error
	f := &File{
		context: context,
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

func (f *File) Context() *Context {
	return f.context
}

func (f *File) Parent() *Dir {
	return f.parent
}

func (f *File) SetParent(p *Dir) {
	f.parent = p
}

func (f *File) Dirname() string {
	return f.dirname
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

// Non interface methods

func (f *File) line(depth int) []byte {
	var indent, prefix, delimiter, name string
	if depth > 0 {
		indent = strings.Repeat(f.context.Config.Indent, depth-1)
		if f.selected {
			prefix = f.context.Config.PrefixSelected
		} else {
			prefix = f.context.Config.PrefixFile
		}
		delimiter = " "
	}
	name = OriginalPath(f)
	return []byte(indent + prefix + delimiter + name)
}

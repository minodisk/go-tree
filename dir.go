package tree

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	OpenedPrefix   = "+"
	ClosedPrefix   = "-"
	SelectedPrefix = "|"
)

type Dir struct {
	os.FileInfo
	config  Config
	dirname string
	parent  *Dir

	selected bool
	opened   bool
	children Operators
}

func NewDir(path string, config Config) (*Dir, error) {
	var err error
	d := &Dir{
		config:  config,
		dirname: filepath.Dir(path),
	}
	d.FileInfo, err = os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !d.FileInfo.IsDir() {
		return nil, fmt.Errorf("the path '%s' isn't directory", path)
	}
	return d, err
}

func (d *Dir) Type() string {
	return "directory"
}

func (d *Dir) SetParent(p *Dir) {
	d.parent = p
}

func (d *Dir) Parent() *Dir {
	return d.parent
}

func (d *Dir) Path() string {
	return filepath.Join(d.dirname, d.Name())
}

func (d *Dir) Selected() bool {
	return d.selected
}

func (d *Dir) Select() {
	d.selected = true
}

func (d *Dir) Unselect() {
	d.selected = false
}

func (d *Dir) ToggleSelected() {
	d.selected = !d.selected
}

func (d *Dir) Rename(newName string) error {
	return os.Rename(d.Path(), filepath.Join(d.dirname, newName))
}

func (d *Dir) Move(newDirname string) error {
	return os.Rename(d.Path(), filepath.Join(newDirname, d.Name()))
}

func (d *Dir) Remove() error {
	return os.RemoveAll(d.Path())
}

// Non interface methods

func (d *Dir) Equals(t *Dir) bool {
	if t == nil {
		return false
	}
	return d.Path() == t.Path()
}

func (d *Dir) Scan() error {
	if !d.opened {
		return nil
	}

	olds := d.children

	d.children = Operators{}
	dirname := d.Path()
	infos, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}
	for _, info := range infos {
		var o Operator
		if info.IsDir() {
			newDir := &Dir{FileInfo: info, config: d.config, dirname: dirname}
			oldDir := olds.FindDir(newDir)
			if oldDir != nil {
				oldDir.Scan()
				o = oldDir
			} else {
				o = newDir
			}
		} else {
			newFile := &File{FileInfo: info, config: d.config, dirname: dirname}
			oldFile := olds.FindFile(newFile)
			if oldFile != nil {
				o = oldFile
			} else {
				o = newFile
			}
		}
		d.AppendChild(o)
	}
	sort.Sort(d.children)
	return nil
}

func (d *Dir) OpenRec() error {
	if err := d.Open(); err != nil {
		return err
	}
	for _, o := range d.children {
		c, ok := o.(*Dir)
		if !ok {
			continue
		}
		if err := c.OpenRec(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dir) Opened() bool {
	return d.opened
}

func (d *Dir) Open() error {
	d.opened = true
	return d.Scan()
}

func (d *Dir) Close() {
	d.opened = false
	d.children = Operators{}
}

func (d *Dir) Toggle() error {
	if d.opened {
		d.Close()
		return nil
	}
	return d.Open()
}

func (d *Dir) ToggleRec() error {
	if d.opened {
		d.Close()
		return nil
	}
	return d.OpenRec()
}

func (d *Dir) NumChildren() int {
	return len(d.children)
}

func (d *Dir) AppendChild(o Operator) {
	d.children = append(d.children, o)
	o.SetParent(d)
}

func (d *Dir) IndexOf(i int) (Operator, bool) {
	o, _, ok := d.indexOf(i)
	return o, ok
}

func (d *Dir) indexOf(i int) (Operator, int, bool) {
	if i == 0 {
		return d, i, true
	}
	for _, o := range d.children {
		i--
		if i == 0 {
			return o, i, true
		}
		if t, ok := o.(*Dir); ok {
			var (
				operator Operator
				found    bool
			)
			operator, i, found = t.indexOf(i)
			if found {
				return operator, i, found
			}
		}
	}
	return nil, i, false
}

func (d *Dir) ReadParent() (*Dir, error) {
	if d.parent != nil {
		return d.parent, nil
	}
	if filepath.ToSlash(d.Path()) == "/" {
		return nil, errors.New("can't read parent")
	}
	p, err := NewDir(d.dirname, d.config)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (d *Dir) HasSelected() bool {
	if d.selected {
		return true
	}
	for _, o := range d.children {
		switch o := o.(type) {
		case *Dir:
			if o.HasSelected() {
				return true
			}
		case *File:
			if o.selected {
				return true
			}
		}
	}
	return false
}

func (d *Dir) Selecteds() Operators {
	os := Operators{}
	if d.selected {
		os = append(os, d)
	}
	for _, o := range d.children {
		switch o := o.(type) {
		case *Dir:
			os = append(os, o.Selecteds()...)
		case *File:
			if o.selected {
				os = append(os, o)
			}
		}
	}
	return os
}

func (d *Dir) Lines(depth int) [][]byte {
	lines := [][]byte{
		d.line(depth),
	}
	depth++

	for _, o := range d.children {
		switch o := o.(type) {
		case *Dir:
			lines = append(lines, o.Lines(depth)...)
		case *File:
			lines = append(lines, o.line(depth))
		}
	}
	return lines
}

func (d *Dir) line(depth int) []byte {
	var prefix string
	if d.selected {
		prefix = d.config.PrefixSelected
	} else if d.opened {
		prefix = d.config.PrefixDirOpened
	} else {
		prefix = d.config.PrefixDirClosed
	}
	return []byte(fmt.Sprintf("%s%s %s/", strings.Repeat(d.config.Indent, depth), prefix, d.Name()))
}

func (d *Dir) CreateDir(name ...string) error {
	for _, n := range name {
		if err := os.MkdirAll(filepath.Join(d.Path(), n), 775); err != nil {
			return err
		}
	}
	return d.Scan()
}

func (d *Dir) CreateFile(name ...string) error {
	for _, n := range name {
		if _, err := os.OpenFile(filepath.Join(d.Path(), n), os.O_CREATE, 664); err != nil {
			return err
		}
	}
	return d.Scan()
}

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

func (d *Dir) Parent() *Dir {
	return d.parent
}

func (d *Dir) SetParent(p *Dir) {
	d.parent = p
}

func (d *Dir) Dirname() string {
	return d.dirname
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

// Non interface methods

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

type Range struct {
	Start, End int
}

type ContextOperators struct {
	Operators Operators
	Caret     int
}

func (d *Dir) ObjectsAt(r Range) Operators {
	ctx := ContextOperators{
		Operators: Operators{},
		Caret:     0,
	}
	d.objectsAt(r, &ctx)
	return ctx.Operators
}

func (d *Dir) objectsAt(r Range, ctx *ContextOperators) bool {
	if ctx.Caret > r.End {
		return false
	}
	if r.Start <= ctx.Caret && ctx.Caret <= r.End {
		ctx.Operators = append(ctx.Operators, d)
	}

	for _, o := range d.children {
		ctx.Caret++
		if ctx.Caret > r.End {
			return false
		}

		if o, ok := o.(*Dir); ok {
			if !o.objectsAt(r, ctx) {
				return false
			}
		} else {
			if r.Start <= ctx.Caret && ctx.Caret <= r.End {
				ctx.Operators = append(ctx.Operators, d)
			}
		}
	}
	return true
}

func (d *Dir) IndexOf(i int) (Operator, bool) {
	o, ok, _ := d.indexOf(i)
	return o, ok
}

func (d *Dir) indexOf(i int) (Operator, bool, int) {
	if i == 0 {
		return d, true, i
	}
	for _, o := range d.children {
		i--
		if i == 0 {
			return o, true, i
		}
		if t, ok := o.(*Dir); ok {
			var (
				operator Operator
				found    bool
			)
			operator, found, i = t.indexOf(i)
			if found {
				return operator, found, i
			}
		}
	}
	return nil, false, i
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

func (d *Dir) All() Operators {
	os := Operators{}
	os = append(os, d)
	for _, o := range d.children {
		switch o := o.(type) {
		case *Dir:
			os = append(os, o.All()...)
		case *File:
			os = append(os, o)
		}
	}
	return os
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
		if err := os.MkdirAll(filepath.Join(d.Path(), n), 0775); err != nil {
			return err
		}
	}
	return d.Scan()
}

func (d *Dir) CreateFile(name ...string) error {
	for _, n := range name {
		if _, err := os.OpenFile(filepath.Join(d.Path(), n), os.O_CREATE, 0664); err != nil {
			return err
		}
	}
	return d.Scan()
}

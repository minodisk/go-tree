package tree

import (
	"errors"
	"path/filepath"
)

var (
	ConfigDefault = NewConfig("", "", "", "", "")
)

type Tree struct {
	root           *Dir
	handleOpenFile func(string) error
	handleChanged  func() error
}

type Config struct {
	Indent          string
	PrefixDirOpened string
	PrefixDirClosed string
	PrefixFile      string
	PrefixSelected  string
}

func NewConfig(indent, prefixDirOpened, prefixDirClosed, prefixFile, prefixSelected string) Config {
	if indent == "" {
		indent = " "
	}
	if prefixDirOpened == "" {
		prefixDirOpened = "-"
	}
	if prefixDirClosed == "" {
		prefixDirClosed = "+"
	}
	if prefixFile == "" {
		prefixFile = "|"
	}
	if prefixSelected == "" {
		prefixSelected = "*"
	}
	return Config{
		indent,
		prefixDirOpened,
		prefixDirClosed,
		prefixFile,
		prefixSelected,
	}
}

func New(path string, config Config) (*Tree, error) {
	var err error
	t := &Tree{}
	t.root, err = NewDir(path, config)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Tree) IndexOf(i int) (Operator, bool) {
	return t.root.IndexOf(i)
}

func (t *Tree) HandleOpenFile(fn func(string) error) {
	t.handleOpenFile = fn
}

func (t *Tree) HandleChanged(fn func() error) {
	t.handleChanged = fn
}

func (t *Tree) OnOpenFile(path string) error {
	if t.handleOpenFile == nil {
		return nil
	}
	return t.handleOpenFile(path)
}

func (t *Tree) Lines() [][]byte {
	return t.root.Lines(0)
}

func (t *Tree) Open() error {
	return t.root.Open()
}

func (t *Tree) UpAt(cursor int) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}

	var current *Dir
	switch o := o.(type) {
	case *Dir:
		if o.opened {
			current = o
		} else {
			var err error
			current, err = o.ReadParent()
			if err != nil {
				return err
			}
		}
	case *File:
		current = o.Parent()
	}
	next, err := current.ReadParent()
	if err != nil {
		return err
	}

	// When target directory is lower than or equal to root,
	// just close current directory.
	isUnder, err := UnderOrEquals(t.root, next)
	if err != nil {
		return err
	}
	if isUnder {
		current.Close()
		return nil
	}

	// The other cases, set target directory as root.
	t.root = next
	if err := t.root.Open(); err != nil {
		return err
	}
	return nil
}

func (t *Tree) DownAt(cursor int) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}

	switch o := o.(type) {
	case *Dir:
		t.root = o
		return t.root.Open()
	case *File:
		return t.OnOpenFile(o.Path())
	default:
		return errors.New("invalid operator")
	}
}

func (t *Tree) ToggleAt(cursor int) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	switch o := o.(type) {
	case *Dir:
		return o.Toggle()
	case *File:
		return o.Parent().Toggle()
	default:
		return errors.New("invalid operator")
	}
}

func (t *Tree) ToggleRecAt(cursor int) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	switch o := o.(type) {
	case *Dir:
		return o.ToggleRec()
	case *File:
		return o.Parent().ToggleRec()
	default:
		return errors.New("invalid operator")
	}
}

func (t *Tree) CreateDirAt(cursor int, names ...string) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	switch o := o.(type) {
	case *Dir:
		return o.CreateDir(names...)
	case *File:
		return o.Parent().CreateDir(names...)
	default:
		return errors.New("invalid operator")
	}
}

func (t *Tree) CreateFileAt(cursor int, names ...string) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	switch o := o.(type) {
	case *Dir:
		return o.CreateFile(names...)
	case *File:
		return o.Parent().CreateFile(names...)
	default:
		return errors.New("invalid operator")
	}
}

// func (t *Tree) NameAt(cursor int) (string, error) {
// 	o, ok := t.root.IndexOf(cursor)
// 	if !ok {
// 		//
// 	}
// 	return o.Name(), nil
// }

// func (t *Tree) RenameAt(cursor int, name string) error {
// 	o, ok := t.root.IndexOf(cursor)
// 	if !ok {
// 		//
// 	}
// 	return o.Rename(name)
// }

func (t *Tree) ToggleSelectedAt(cursor int) (bool, error) {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	o.ToggleSelected()
	return o.Selected(), nil
}

func (t *Tree) HasSelected() bool {
	return t.root.HasSelected()
}

func (t *Tree) RenameSelected() error {
	os := t.root.Selecteds()
	for _, o := range os {
		o.Unselect()
	}
	return errors.New("rename of multiple files has not been implemented yet")
}

func (t *Tree) MoveSelected(path string) error {
	os := t.root.Selecteds()
	for _, o := range os {
		o.Unselect()
	}
	for _, o := range os {
		if err := o.Move(filepath.Join(t.root.Path(), path)); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) MoveAt(cursor int, name string) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	return o.Move(filepath.Join(t.root.Path(), name))
}

func (t *Tree) RemoveSelected() error {
	os := t.root.Selecteds()
	for _, o := range os {
		o.Unselect()
	}
	for _, o := range os {
		if err := o.Remove(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) RemoveAt(cursor int, name string) error {
	o, ok := t.root.IndexOf(cursor)
	if !ok {
		//
	}
	return o.Remove()
}

func (t *Tree) Scan() error {
	return t.root.Scan()
}

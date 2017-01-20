package tree

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/mattn/natural"
	"github.com/skratchdot/open-golang/open"
)

type Operator interface {
	SetParent(*Dir)
	Parent() *Dir
	IsDir() bool
	Type() string
	Name() string
	Dirname() string
	Path() string

	Selected() bool
	Select()
	Unselect()
	ToggleSelected()

	Move(string) error
	Remove() error
}

func Equals(a, b Operator) bool {
	if a.Type() != b.Type() {
		return false
	}
	return a.Path() == b.Path()
}

func Rel(base, target Operator) (string, error) {
	return filepath.Rel(base.Path(), target.Path())
}

func UnderOrEquals(base, target Operator) (bool, error) {
	rel, err := Rel(base, target)
	if err != nil {
		return false, err
	}
	l := len(rel)
	switch {
	case l < 2:
		return true, nil
	case l == 2:
		return rel != "..", nil
	default:
		return rel[0:3] != "../", nil
	}
}

func GetDir(o Operator) *Dir {
	switch t := o.(type) {
	case *Dir:
		return t
	default:
		return t.Parent()
	}
}

func GetDirWithOpened(o Operator) (*Dir, error) {
	switch o := o.(type) {
	case *Dir:
		if o.opened {
			return o, nil
		}
		return o.ReadParent()
	default:
		return o.Parent(), nil
	}
}

func Toggle(o Operator) error {
	switch o := o.(type) {
	case *Dir:
		return o.Toggle()
	case *File:
		return o.Parent().Toggle()
	default:
		return errors.New("invalid operator")
	}
}

func ToggleRec(o Operator) error {
	switch o := o.(type) {
	case *Dir:
		return o.ToggleRec()
	case *File:
		return o.Parent().ToggleRec()
	default:
		return errors.New("invalid operator")
	}
}

func CreateDir(o Operator, names ...string) error {
	switch o := o.(type) {
	case *Dir:
		return o.CreateDir(names...)
	case *File:
		return o.Parent().CreateDir(names...)
	default:
		return errors.New("invalid operator")
	}
}

func CreateFile(o Operator, names ...string) error {
	switch o := o.(type) {
	case *Dir:
		return o.CreateFile(names...)
	case *File:
		return o.Parent().CreateFile(names...)
	default:
		return errors.New("invalid operator")
	}
}

func Rename(o Operator, newName string) error {
	return os.Rename(o.Path(), filepath.Join(o.Dirname(), newName))
}

func OpenWithOS(o Operator) error {
	return open.Run(o.Path())
}

type Operators []Operator

func (os Operators) Len() int {
	return len(os)
}

func (os Operators) Swap(i, j int) {
	os[i], os[j] = os[j], os[i]
}

func (os Operators) Less(i, j int) bool {
	a, b := os[i], os[j]
	aIsDir, bIsDir := a.IsDir(), b.IsDir()
	if aIsDir && !bIsDir {
		return true
	}
	if !aIsDir && bIsDir {
		return false
	}
	return natural.NaturalComp(a.Name(), b.Name()) < 0
}

func (os Operators) Select() {
	for _, o := range os {
		o.Select()
	}
}

func (os Operators) Unselect() {
	for _, o := range os {
		o.Unselect()
	}
}

func (os Operators) FindDir(d *Dir) *Dir {
	for _, o := range os {
		t, ok := o.(*Dir)
		if !ok {
			continue
		}
		if t.Equals(d) {
			return t
		}
	}
	return nil
}

func (os Operators) FindFile(f *File) *File {
	for _, o := range os {
		t, ok := o.(*File)
		if !ok {
			continue
		}
		if t.Equals(f) {
			return t
		}
	}
	return nil
}

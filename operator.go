package tree

import "path/filepath"

type Operator interface {
	SetParent(*Dir)
	Parent() *Dir
	IsDir() bool
	Type() string
	Name() string
	Path() string

	Selected() bool
	Select()
	Unselect()
	ToggleSelected()

	Rename(string) error
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
	return a.Name() < b.Name()
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

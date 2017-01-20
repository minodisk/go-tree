package tree

import "github.com/mattn/natural"

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
		if Equals(t, d) {
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
		if Equals(t, f) {
			return t
		}
	}
	return nil
}

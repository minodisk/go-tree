package tree

type Operator interface {
	setParent(*Dir)
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

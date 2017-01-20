package tree

import (
	"os"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
)

// A Operator represents objects in file system.
type Operator interface {
	IsDir() bool
	Name() string
	Dirname() string
	Path() string
	Parent() *Dir
	SetParent(*Dir)
	Selected() bool
	Select()
	Unselect()
	ToggleSelected()
}

// Type returns the type of Operator.
func Type(o Operator) string {
	switch o.(type) {
	case *Dir:
		return "directory"
	case *File:
		return "file"
	default:
		return "undefined"
	}
}

// Equals returns that two Operators are pointing same object.
func Equals(a, b Operator) bool {
	if a.IsDir() != b.IsDir() {
		return false
	}
	return a.Path() == b.Path()
}

// Rel returns relative path of two Operators.
func Rel(base, target Operator) (string, error) {
	return filepath.Rel(base.Path(), target.Path())
}

// UnderOrEquals returns that the path of target object is under or equals to the path of base directory.
func UnderOrEquals(base *Dir, target Operator) (bool, error) {
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

// GetDir returns nearest directory.
// When o is *Dir, returns itself.
// In other cases, returns the parent of o.
func GetDir(o Operator) *Dir {
	switch t := o.(type) {
	case *Dir:
		return t
	default:
		return t.Parent()
	}
}

// GetDirWithOpened return nearest directory.
// The difference from GetDir appears in case that o is *Dir.
// When the *Dir is opened, returns itself.
// When the *Dir is closed, returns the parent of the *Dir.
// In other cases, returns the parent of o.
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

// Toggle toggles opened status.
// When o is *Dir, toggles opened status of itself.
// In other cases, toggles opened status of the parent of o.
func Toggle(o Operator) error {
	switch o := o.(type) {
	case *Dir:
		return o.Toggle()
	default:
		return o.Parent().Toggle()
	}
}

// ToggleRec toggles opened status recursively.
// When o is *Dir, toggles opened status recursively under itself.
// In other cases, toggles opened status recursively under the parent of o.
func ToggleRec(o Operator) error {
	switch o := o.(type) {
	case *Dir:
		return o.ToggleRec()
	default:
		return o.Parent().ToggleRec()
	}
}

// CreateDir makes new directories.
// When o is *Dir, makes under itself.
// In other cases, makes under the parent of o.
func CreateDir(o Operator, names ...string) error {
	switch o := o.(type) {
	case *Dir:
		return o.CreateDir(names...)
	default:
		return o.Parent().CreateDir(names...)
	}
}

// CreateFile makes new files.
// When o is *Dir, makes under itself.
// In other cases, makes under the parent of o.
func CreateFile(o Operator, names ...string) error {
	switch o := o.(type) {
	case *Dir:
		return o.CreateFile(names...)
	default:
		return o.Parent().CreateFile(names...)
	}
}

// Rename renames o to newName.
func Rename(o Operator, newName string) error {
	return os.Rename(o.Path(), filepath.Join(o.Dirname(), newName))
}

// Move moves o to under the newDirname.
func Move(o Operator, newDirname string) error {
	return os.Rename(o.Path(), filepath.Join(newDirname, o.Name()))
}

// Remove removes o and any children it contains.
func Remove(o Operator) error {
	return os.RemoveAll(o.Path())
}

// OpenWithOS opens o with the default application related in OS.
func OpenWithOS(o Operator) error {
	return open.Run(o.Path())
}

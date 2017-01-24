package tree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	shutil "github.com/termie/go-shutil"
)

type CursorFunc func() (int, error)
type SetCursorFunc func(int) error
type ConfirmFunc func(...Operator) (bool, error)
type TextFunc func() (string, error)
type TextsFunc func() ([]string, error)
type OperatorFunc func(Operator) error
type OperatorsFunc func(Operators) error
type OperatorTextFunc func(Operator) (string, error)
type OperatorsTextFunc func(Operators) (string, error)
type OperatorsTextsFunc func(Operators) ([]string, error)
type OpenFileFunc func(*File) error
type CancelFunc func() error
type RenderFunc func([][]byte) error
type SetClipboardFunc func(string) error

type Tree struct {
	root    *Dir
	context *Context
}

func New(path string, context *Context) (*Tree, error) {
	if err := context.Init(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(context.Config.TrashDirname, 0775); err != nil {
		return nil, err
	}

	t := &Tree{context: context}
	if err := t.SetRootPath(path); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Tree) SetRootPath(path string) error {
	root, err := NewDir(path, t.context)
	if err != nil {
		return err
	}
	return t.SetRoot(root)
}

func (t *Tree) SetRoot(root *Dir) error {
	t.root = root
	return t.root.Open()
}

func (t *Tree) Operator(cursor CursorFunc) (Operator, error) {
	c, err := cursor()
	if err != nil {
		return nil, err
	}
	o, ok := t.root.IndexOf(c)
	if !ok {
		//
	}
	return o, nil
}

type SelectedRangeFunc func() (Range, error)

func (t *Tree) Operators(selectedRange SelectedRangeFunc) (Operators, error) {
	r, err := selectedRange()
	if err != nil {
		return nil, err
	}
	return t.root.ObjectsAt(r), nil
}

func (t *Tree) IndexOf(i int) (Operator, bool) {
	return t.root.IndexOf(i)
}

func (t *Tree) HasSelected() bool {
	return t.root.HasSelected()
}

func (t *Tree) Render(render RenderFunc) error {
	return render(t.root.Lines(0))
}

func (t *Tree) ScanAndRender(render RenderFunc) error {
	t.root.Scan()
	return t.Render(render)
}

func (t *Tree) Open(render RenderFunc) error {
	defer t.Render(render)
	return t.root.Open()
}

func (t *Tree) CD(path TextFunc, render RenderFunc) error {
	defer t.Render(render)
	p, err := path()
	if err != nil {
		return err
	}
	return t.SetRootPath(p)
}

func (t *Tree) Root(render RenderFunc) error {
	defer t.Render(render)
	return t.SetRootPath(DirRoot())
}

func (t *Tree) Home(render RenderFunc) error {
	defer t.Render(render)
	dir, err := DirHome()
	if err != nil {
		return err
	}
	return t.SetRootPath(dir)
}

func (t *Tree) Trash(render RenderFunc) error {
	defer t.Render(render)
	return t.SetRootPath(t.context.Config.TrashDirname)
}

func (t *Tree) Project(render RenderFunc) error {
	defer t.Render(render)

	p, err := DirProject(t.root.Path(), t.context.Config.rProject)
	if err != nil {
		return err
	}
	return t.SetRootPath(p)
}

func (t *Tree) Up(cursor CursorFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}

	current, err := NearestOpenedDir(o)
	if err != nil {
		return err
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
	return t.root.Open()
}

func (t *Tree) Down(cursor CursorFunc, openFile OpenFileFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}

	switch o := o.(type) {
	case *Dir:
		t.root = o
		return t.root.Open()
	case *File:
		return openFile(o)
	default:
		return errors.New("invalid operator")
	}
}

func (t *Tree) Select(cursor CursorFunc, setCursorFunc SetCursorFunc, render RenderFunc) error {
	defer t.Render(render)

	c, err := cursor()
	if err != nil {
		return err
	}
	o, ok := t.IndexOf(c)
	if !ok {
		//
	}
	o.ToggleSelected()
	if o.Selected() {
		return setCursorFunc(c + 1)
	}
	return nil
}

func (t *Tree) ReverseSelected(render RenderFunc) error {
	for _, o := range t.root.All() {
		if o.Selected() {
			o.Unselect()
		} else {
			o.Select()
		}
	}
	return t.Render(render)
}

func (t *Tree) Toggle(cursor CursorFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	return Toggle(o)
}

func (t *Tree) ToggleRec(cursor CursorFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	return ToggleRec(o)
}

func (t *Tree) CreateDir(cursor CursorFunc, texts TextsFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	names, err := texts()
	if err != nil {
		return err
	}
	return CreateDir(o, names...)
}

func (t *Tree) CreateFile(cursor CursorFunc, texts TextsFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	names, err := texts()
	if err != nil {
		return err
	}
	return CreateFile(o, names...)
}

func (t *Tree) Rename(cursor CursorFunc, text OperatorTextFunc, texts OperatorsTextsFunc, cancel CancelFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		names, err := texts(os)
		if err != nil {
			return err
		}
		before := len(os)
		after := len(names)
		if after != before {
			return fmt.Errorf("the number of names differs before(%d) and after(%d)", before, after)
		}
		for i, o := range os {
			n := names[i]
			if err := Rename(o, n); err != nil {
				return err
			}
		}
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	new, err := text(o)
	if err != nil {
		return err
	}
	return Rename(o, new)
}

func (t *Tree) Move(cursor CursorFunc, text OperatorsTextFunc, cancel CancelFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		path, err := text(os)
		if err != nil {
			return err
		}
		if path == "" {
			return cancel()
		}
		for _, o := range os {
			if err := Move(o, filepath.Join(t.root.Path(), path)); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	path, err := text(Operators{o})
	if err != nil {
		return err
	}
	return Move(o, filepath.Join(t.root.Path(), path))
}

func (t *Tree) Remove(cursor CursorFunc, confirm ConfirmFunc, cancel CancelFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		ok, err := confirm(os...)
		if err != nil {
			return err
		}
		if !ok {
			return cancel()
		}
		for _, o := range os {
			if err := Remove(o); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	ok, err := confirm(o)
	if err != nil {
		return err
	}
	if !ok {
		return cancel()
	}
	return Remove(o)
}

func (t *Tree) RemovePermanently(cursor CursorFunc, confirm ConfirmFunc, cancel CancelFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		ok, err := confirm(os...)
		if err != nil {
			return err
		}
		if !ok {
			return cancel()
		}
		for _, o := range os {
			if err := RemovePermanently(o); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	ok, err := confirm(o)
	if err != nil {
		return err
	}
	if !ok {
		return cancel()
	}
	return RemovePermanently(o)
}

func (t *Tree) Restore(cursor CursorFunc, confirm ConfirmFunc, cancel CancelFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		ok, err := confirm(os...)
		if err != nil {
			return err
		}
		if !ok {
			return cancel()
		}
		for _, o := range os {
			if err := Restore(o); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	ok, err := confirm(o)
	if err != nil {
		return err
	}
	if !ok {
		return cancel()
	}
	return Restore(o)
}

func (t *Tree) OpenExternally(cursor CursorFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		for _, o := range os {
			if err := OpenWithOS(o); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	return OpenWithOS(o)
}

func (t *Tree) OpenDirExternally(cursor CursorFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		for _, o := range os {
			if err := OpenWithOS(NearestDir(o)); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	return OpenWithOS(NearestDir(o))
}

func (t *Tree) Copy(cursor CursorFunc) error {
	if t.HasSelected() {
		os := t.root.Selecteds()
		defer os.Unselect()
		t.context.Registry = os
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	t.context.Registry = Operators{o}
	return nil
}

func (t *Tree) CopiedList(operators OperatorsFunc) error {
	return operators(t.context.Registry)
}

func (t *Tree) Paste(cursor CursorFunc, render RenderFunc) error {
	defer t.ScanAndRender(render)

	if t.context.Registry.Len() == 0 {
		return nil
	}
	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	d, err := NearestOpenedDir(o)
	if err != nil {
		return err
	}
	dst := d.Path()
	for _, o := range t.context.Registry {
		if o.IsDir() {
			if err := shutil.CopyTree(o.Path(), dst, nil); err != nil {
				return err
			}
		} else {
			if err := shutil.CopyFile(o.Path(), filepath.Join(dst, o.Name()), true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Tree) Yank(cursor CursorFunc, setClipboard SetClipboardFunc) error {
	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	return setClipboard(o.Path())
}

package tree

import (
	"errors"
	"fmt"
	"path/filepath"
)

type CursorFunc func() (int, error)
type SetCursorFunc func(int) error
type ConfirmFunc func(...Operator) (bool, error)
type TextFunc func() (string, error)
type TextsFunc func() ([]string, error)
type OperatorTextFunc func(Operator) (string, error)
type OperatorsTextFunc func(...Operator) (string, error)
type OperatorsTextsFunc func(...Operator) ([]string, error)
type OpenFileFunc func(*File) error
type CancelFunc func() error
type RenderFunc func([][]byte) error

type Tree struct {
	root   *Dir
	config Config
}

func New(path string, config Config) (*Tree, error) {
	t := &Tree{config: config}
	if err := t.SetRootPath(path); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Tree) SetRootPath(path string) error {
	root, err := NewDir(path, t.config)
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

func (t *Tree) Up(cursor CursorFunc, render RenderFunc) error {
	defer t.Render(render)

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}

	current, err := GetDirWithOpened(o)
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

func (t *Tree) SelectAll(render RenderFunc) error {
	if t.HasSelected() {
		t.root.Selecteds().Unselect()
		return t.Render(render)
	}
	t.root.All().Select()
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
		names, err := texts(os...)
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
		path, err := text(os...)
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
	path, err := text(o)
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
			if err := OpenWithOS(GetDir(o)); err != nil {
				return err
			}
		}
		return nil
	}

	o, err := t.Operator(cursor)
	if err != nil {
		return err
	}
	return OpenWithOS(GetDir(o))
}

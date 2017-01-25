package tree_test

import (
	"path/filepath"
	"testing"

	tree "github.com/minodisk/go-tree"
)

type O struct {
	path string
}

func NewO(path string) *O           { return &O{path: path} }
func (o *O) Context() *tree.Context { return &tree.Context{} }
func (o *O) SetParent(d *tree.Dir)  {}
func (o *O) Parent() *tree.Dir      { return &tree.Dir{} }
func (o *O) IsDir() bool            { return false }
func (o *O) Type() string           { return "" }
func (o *O) Path() string           { return o.path }
func (o *O) Dirname() string        { return filepath.Dir(o.path) }
func (o *O) Name() string           { return filepath.Base(o.path) }
func (o *O) Selected() bool         { return false }
func (o *O) Select()                {}
func (o *O) Unselect()              {}
func (o *O) ToggleSelected()        {}
func (o *O) Rename(string) error    { return nil }
func (o *O) Move(string) error      { return nil }
func (o *O) Remove() error          { return nil }

func TestRel(t *testing.T) {
	type Case struct {
		Base     *O
		Target   *O
		Expected string
	}
	for _, c := range []Case{
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo"),
			Expected: ".",
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/bar"),
			Expected: "bar",
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/.bar"),
			Expected: ".bar",
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/..bar"),
			Expected: "..bar",
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/.bar/baz"),
			Expected: ".bar/baz",
		},
		{
			Base:     NewO("/foo/bar"),
			Target:   NewO("/foo"),
			Expected: "..",
		},
		{
			Base:     NewO("/foo/bar/baz"),
			Target:   NewO("/foo/qux"),
			Expected: "../../qux",
		},
		{
			Base:     NewO("/foo/bar/baz"),
			Target:   NewO("/foo/.bar"),
			Expected: "../../.bar",
		},
	} {
		actual, err := tree.Rel(c.Base, c.Target)
		if err != nil {
			t.Fatal(err)
		}
		if actual != c.Expected {
			t.Errorf("Rel(%v, %v) expected %v, but actual %v", c.Base, c.Target, c.Expected, actual)
		}
	}
}

func TestEquals(t *testing.T) {
	type Case struct {
		Base     *O
		Target   *O
		Expected bool
	}
	for _, c := range []Case{
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo"),
			Expected: true,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/bar"),
			Expected: false,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/.bar"),
			Expected: false,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/..bar"),
			Expected: false,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/.bar/baz"),
			Expected: false,
		},
		{
			Base:     NewO("/foo/bar"),
			Target:   NewO("/foo"),
			Expected: false,
		},
		{
			Base:     NewO("/foo/bar/baz"),
			Target:   NewO("/foo/qux"),
			Expected: false,
		},
		{
			Base:     NewO("/foo/bar/baz"),
			Target:   NewO("/foo/.bar"),
			Expected: false,
		},
	} {
		actual := tree.Equals(c.Base, c.Target)
		if actual != c.Expected {
			t.Errorf("Equals(%v, %v) expected %v, but actual %v", c.Base, c.Target, c.Expected, actual)
		}
	}
}

func TestUnderOrEquals(t *testing.T) {
	type Case struct {
		Base     *O
		Target   *O
		Expected bool
	}
	for _, c := range []Case{
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo"),
			Expected: true,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/bar"),
			Expected: true,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/.bar"),
			Expected: true,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/..bar"),
			Expected: true,
		},
		{
			Base:     NewO("/foo"),
			Target:   NewO("/foo/.bar/baz"),
			Expected: true,
		},
		{
			Base:     NewO("/foo/bar"),
			Target:   NewO("/foo"),
			Expected: false,
		},
		{
			Base:     NewO("/foo/bar/baz"),
			Target:   NewO("/foo/qux"),
			Expected: false,
		},
		{
			Base:     NewO("/foo/bar/baz"),
			Target:   NewO("/foo/.bar"),
			Expected: false,
		},
	} {
		actual, err := tree.UnderOrEquals(c.Base, c.Target)
		if err != nil {
			t.Fatal(err)
		}
		if actual != c.Expected {
			t.Errorf("UnderOrEquals(%v, %v) expected %v, but actual %v", c.Base, c.Target, c.Expected, actual)
		}
	}
}

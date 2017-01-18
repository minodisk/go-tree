package tree_test

import (
	"fmt"
	"strings"
	"testing"

	tree "github.com/minodisk/go-tree"
)

func linesToString(lines [][]byte) string {
	ls := make([]string, len(lines))
	for i, line := range lines {
		ls[i] = string(line)
	}
	return strings.Join(ls, "\n")
}

func TestNewDir(t *testing.T) {
	{
		_, err := tree.NewDir("/home/minodisk/Workspace/go/src/github.com/minodisk/nvim-finder", tree.ConfigDefault)
		if err != nil {
			t.Errorf("success with nvim-finder directory")
		}
	}

	{
		_, err := tree.NewDir("./fixtures/wrong/path", tree.ConfigDefault)
		if err == nil {
			t.Errorf("should return error with a wrong path")
		}
	}

	{
		_, err := tree.NewDir("./fixtures/foo/a.txt", tree.ConfigDefault)
		if err == nil {
			t.Errorf("should return error with a file path")
		}
	}

	{
		d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
		if err != nil {
			t.Fatalf("fail to create dir with error: %s", err)
		}
		a := linesToString(d.Lines(0))
		e := `+ foo/`
		if a != e {
			t.Errorf("NewDir().Lines() should be\nexpected:\n%s\nactual:\n%s", e, a)
		}
	}
}

func TestType(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if d.Type() != "directory" {
		t.Errorf("Type() should return directory")
	}
}

func TestOpenRec(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.OpenRec(); err != nil {
		t.Fatal(err)
	}
	a := linesToString(d.Lines(0))
	e := `- foo/
 - bar/
  - baz/
   - qux/
 | a.txt
 | b.txt
 | c.txt`
	if a != e {
		t.Errorf("OpenRec().Lines() should be\nexpected:\n%s\nactual:\n%s", e, a)
	}
}

func TestOpened(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if d.Opened() {
		t.Errorf("created Dir shouldn't be opened")
	}
	if err := d.Open(); err != nil {
		t.Fatal(err)
	}
	if !d.Opened() {
		t.Errorf("opened Dir should be opened")
	}
	d.Close()
	if d.Opened() {
		t.Errorf("closed Dir shouldn't be opened")
	}
}

func TestOpen(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Open(); err != nil {
		t.Fatal(err)
	}
	a := linesToString(d.Lines(0))
	e := `- foo/
 + bar/
 | a.txt
 | b.txt
 | c.txt`
	if a != e {
		t.Errorf("Open().Lines() should be\nexpected:\n%s\nactual:\n%s", e, a)
	}
}

func TestClose(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Open(); err != nil {
		t.Fatal(err)
	}
	d.Close()
	a := linesToString(d.Lines(0))
	e := `+ foo/`
	if a != e {
		t.Errorf("Close().Lines() should be\nexpected:\n%s\nactual:\n%s", e, a)
	}
}

func TestToggle(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	{
		a := linesToString(d.Lines(0))
		e := `+ foo/`
		if a != e {
			t.Errorf("Before Toggle(), Lines() should return\nexpected:\n%s\nactual:\n%s", e, a)
		}
	}
	if err := d.Toggle(); err != nil {
		t.Fatal(err)
	}
	{
		a := linesToString(d.Lines(0))
		e := `- foo/
 + bar/
 | a.txt
 | b.txt
 | c.txt`
		if a != e {
			t.Errorf("After Toggle() 1st time, Lines() should return\nexpected:\n%s\nactual:\n%s", e, a)
		}
	}
	if err := d.Toggle(); err != nil {
		t.Fatal(err)
	}
	{
		a := linesToString(d.Lines(0))
		e := `+ foo/`
		if a != e {
			t.Errorf("After Toggle() 2nd time, Lines() should return\nexpected:\n%s\nactual:\n%s", e, a)
		}
	}
}

func TestNumChildren(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Open(); err != nil {
		t.Fatal(err)
	}
	a := d.NumChildren()
	e := 4
	if a != e {
		t.Errorf("NumChildren() should count the length of the children")
	}
}

func TestAppendChild(t *testing.T) {
	p, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	c, err := tree.NewDir("./fixtures/foo/bar", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if p.NumChildren() == 0 {
		fmt.Println("Before AppendChild(), children length should be 0")
	}
	if c.Parent() != nil {
		fmt.Println("Before AppendChild(), parent should be nil")
	}
	p.AppendChild(c)
	if p.NumChildren() == 1 {
		fmt.Println("AppendChild() should append child to parent's children")
	}
	if c.Parent() != p {
		fmt.Println("AppendChild() should set parent as child's parent")
	}
}

func TestIndexOf(t *testing.T) {
	d, err := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.OpenRec(); err != nil {
		t.Fatal(err)
	}
	for i, e := range []string{
		"foo",
		"bar",
		"baz",
		"qux",
		"a.txt",
		"b.txt",
		"c.txt",
	} {
		actual, ok := d.IndexOf(i)
		if !ok {
			t.Errorf("IndexOf(%d) should return Operator '%s'", i, e)
			continue
		}
		a := actual.Name()
		if a != e {
			t.Errorf("IndexOf(%d) should return Operator '%s', but actually return '%s'", i, e, a)
		}
	}
}

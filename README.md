# tree [![GoDoc](https://godoc.org/github.com/minodisk/go-tree?status.png)](https://godoc.org/github.com/minodisk/go-tree)

Output directory tree structure, and manipulate the objects in it.

## Installation

```
go get github.com/minodisk/go-tree
```

## Usage

```go
func main() {
	foo, _ := tree.NewDir("./fixtures/foo", tree.ConfigDefault)
	foo.OpenRec()
	foo.Lines() // -> - foo/
	            //     - bar/
	            //      - baz/
	            //       - qux/
	            //     | a.txt
	            //     | b.txt
	            //     | c.txt`

	baz, _ := foo.IndexOf(2)
	baz.Close()
	baz.Lines() // -> - foo/
	            //     - bar/
	            //      + baz/
	            //     | a.txt
	            //     | b.txt
	            //     | c.txt`
}
```

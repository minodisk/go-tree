package tree

import (
	"os/user"
	"path/filepath"
	"regexp"
)

var (
	ConfigDefault = &Config{
		Indent:          " ",
		PrefixDirOpened: "-",
		PrefixDirClosed: "+",
		PrefixFile:      "|",
		PrefixSelected:  "*",
		PostfixDir:      "/",
		RegexpProject:   `^(?:\.git)$`,
	}
)

func init() {
	if err := func() error {
		u, err := user.Current()
		if err != nil {
			return err
		}
		ConfigDefault.TrashDirname = filepath.Join(u.HomeDir, ".finder-trash")
		return nil
	}(); err != nil {
		panic(err)
	}
}

type Context struct {
	Config   *Config
	Registry Operators
}

func (c *Context) Init() error {
	if c.Config == nil {
		c.Config = &Config{}
	}
	c.Config.FillWithDefault()
	return c.Config.Compile()
}

type Config struct {
	Indent          string
	PrefixDirOpened string
	PrefixDirClosed string
	PrefixFile      string
	PrefixSelected  string
	PostfixDir      string
	TrashDirname    string
	RegexpProject   string

	rProject *regexp.Regexp
}

func (c *Config) FillWithDefault() {
	if c.Indent == "" {
		c.Indent = ConfigDefault.Indent
	}
	if c.PrefixDirOpened == "" {
		c.PrefixDirOpened = ConfigDefault.PrefixDirOpened
	}
	if c.PrefixDirClosed == "" {
		c.PrefixDirClosed = ConfigDefault.PrefixDirClosed
	}
	if c.PrefixFile == "" {
		c.PrefixFile = ConfigDefault.PrefixFile
	}
	if c.PrefixSelected == "" {
		c.PrefixSelected = ConfigDefault.PrefixSelected
	}
	if c.PostfixDir == "" {
		c.PostfixDir = ConfigDefault.PostfixDir
	}
	if c.TrashDirname == "" {
		c.TrashDirname = ConfigDefault.TrashDirname
	}
	if c.RegexpProject == "" {
		c.RegexpProject = ConfigDefault.RegexpProject
	}
}

func (c *Config) Compile() error {
	var err error
	c.rProject, err = regexp.Compile(c.RegexpProject)
	return err
}

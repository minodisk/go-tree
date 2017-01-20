package tree

var (
	ConfigDefault = Config{
		Indent:          "  ",
		PrefixDirOpened: "-",
		PrefixDirClosed: "+",
		PrefixFile:      "|",
		PrefixSelected:  "*",
	}
)

type Config struct {
	Indent          string
	PrefixDirOpened string
	PrefixDirClosed string
	PrefixFile      string
	PrefixSelected  string
}

func (c Config) FillWithDefault() Config {
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
	return c
}

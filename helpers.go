package tree

import (
	"io/ioutil"
	"os/user"
	"path/filepath"
	"regexp"
)

func DirRoot() string {
	return "/"
}

func DirHome() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}

func DirProject(dirname string, rProject *regexp.Regexp) (string, error) {
	is, err := ioutil.ReadDir(dirname)
	if err != nil {
		return "", err
	}
	for _, i := range is {
		if rProject.MatchString(i.Name()) {
			return dirname, nil
		}
	}
	parent := filepath.Join(dirname, "..")
	if parent == dirname {
		return dirname, nil
	}
	return DirProject(parent, rProject)
}

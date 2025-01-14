package binds

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type finder struct {
	m map[string]string
}

func newFinder() (*finder, error) {
	var f = &finder{m: map[string]string{}}

	if err := filepath.WalkDir("./", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".go" {
			return err
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return errors.WithStack(err)
		}

		name := filepath.Base(path)
		name = strings.TrimSuffix(name, ".go")
		f.m[name] = string(b)
		return nil
	}); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *finder) FindDeclare(typ string) (filename string, err error) {
	str := fmt.Sprintf("type %s ", typ)
	return f.find(str)
}

func (f *finder) FindMethod(typ, method string) (filename string, err error) {
	str := fmt.Sprintf("%s) %s(", typ, method)
	return f.find(str)
}

func (f *finder) find(s string) (filename string, err error) {
	for n, b := range f.m {
		if strings.Contains(b, s) {
			if filename == "" {
				filename = n
			} else {
				return "", errors.Errorf("find multiple literal %s", s)
			}
		}
	}
	if filename == "" {
		return filename, errors.Errorf("not find literal %s", s)
	} else {
		return filename, nil
	}
}

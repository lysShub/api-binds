package main

import (
	"embed"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//go:embed binds
var code embed.FS

func main() {
	if err := Bind(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	} else {
		os.Remove("gen_test.go")
	}
}

func Bind() error {
	var name = flag.String("name", "", "bind type name")
	var kind = flag.String("kind", "", "bind kind: Gin, Std")
	flag.Parse()
	if *name == "" {
		return errors.New("require parameter name")
	} else {
		// todo 验证类型存在
	}
	switch *kind {
	case "Gin":
	case "Std":
		return errors.Errorf("currently not supported Std kind")
	default:
		return errors.Errorf("invalid kind %s", *kind)
	}

	var pkgname string
	if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return err
		}

		f, err := parser.ParseFile(token.NewFileSet(), d.Name(), nil, parser.PackageClauseOnly)
		if err != nil {
			return errors.WithStack(err)
		}
		pkgname = f.Name.Name
		return filepath.SkipAll
	}); err != nil {
		return err
	}

	s, err := gen_test_func(pkgname, *name, *kind)
	if err != nil {
		return err
	}
	fh, err := os.Create("gen_test.go")
	if err != nil {
		return errors.WithStack(err)
	}
	defer fh.Close()
	if _, err := fh.WriteString(s); err != nil {
		return errors.WithStack(err)
	}

	b, err := exec.Command("go", "mod", "tidy").CombinedOutput()
	if err != nil {
		return errors.WithMessage(err, string(b))
	}

	b, err = exec.Command("go", "test", "-run", fmt.Sprintf("TestRunBind%s", *kind)).CombinedOutput()
	if err != nil {
		return errors.WithMessage(err, string(b))
	}
	return nil
}

func gen_test_func(pkgname, name, kind string) (s string, err error) {
	cmd, vcs, err := command()
	if err != nil {
		return "", err
	}

	fs, err := code.ReadDir("binds")
	if err != nil {
		return "", errors.WithStack(err)
	}

	var l = []io.Reader{test_template(pkgname, name, kind, cmd, vcs)}
	for _, e := range fs {
		f, err := code.Open(path.Join("binds", e.Name()))
		if err != nil {
			return "", errors.WithStack(err)
		}
		l = append(l, f)
	}

	return merge(l, pkgname)
}

func command() (cmd, vcs string, err error) {
	args := os.Args
	p, err := exec.LookPath(args[0])
	if err != nil {
		p = args[0]
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", "", errors.Errorf("cannot debug.ReadBuildInfo from %s", p)
	}

	var v string
	for _, e := range info.Settings {
		if e.Key == "vcs.revision" {
			v = strings.TrimSpace(e.Value)
		}
	}

	if v == "" {
		vcs = info.Path
	} else {
		vcs = fmt.Sprintf("%s@v%s", info.Path, v)
	}
	return strings.Join(args, " "), vcs, nil
}

func test_template(pkgname, name, kind, cmd, vcs string) io.Reader {
	var s = fmt.Sprintf(
		`package %s
import (
	"testing"
	"github.com/stretchr/testify/require"
)
func init() {
	cmd=%s
	vcs=%s
}
func TestRunBind%s(t *testing.T) {
	typ:=reflect.TypeOf(%s{})
	require.NoError(t, %s(typ))
}
`,
		pkgname, strconv.Quote(cmd), strconv.Quote(vcs), kind, name, kind,
	)

	return strings.NewReader(s)
}

package binds

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"

	_ "net/http"

	_ "github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Gin bind gin
func Gin(t reflect.Type) error {
	var b = &builder{
		imports: []string{
			`"github.com/gin-gonic/gin"`,
			`"net/http"`,
		},
	}
	if t.Kind() != reflect.Pointer {
		t = reflect.PointerTo(t)
	}
	b.name = t.Elem().Name()
	if b.name == "" {
		return errors.New("not support anonymous type")
	} else if isuper(b.name[0]) {
		return errors.New("not support exported type")
	}

	f, err := newFinder()
	if err != nil {
		return err
	}
	if declarefile, err := f.FindDeclare(b.name); err != nil {
		return err
	} else {
		f := b.file(declarefile)
		f.typedef = true
	}

	n := t.NumMethod()
	if n == 0 {
		return errors.Errorf("type %s has't method", b.name)
	}
	for i := range n {
		m := t.Method(i)
		file, err := f.FindMethod(b.name, m.Name)
		if err != nil {
			return err
		}
		f := b.file(file)

		name, kind, err := fnname(m.Name)
		if err != nil {
			return err
		} else if name == "" {
			continue
		}

		var e = method{httpmethod: kind, originName: m.Name, name: name}

		if m.Type.NumIn() == 3 {
			for i, t := range []reflect.Type{m.Type.In(1), m.Type.In(2)} {
				if err := checkParaType(t); err != nil {
					return errors.WithMessage(err, m.Name)
				}

				pkgpath, pkgname, err := pkginfo(t)
				if err != nil {
					return err
				}
				imp := strconv.Quote(pkgpath)
				if path.Base(pkgpath) != pkgname {
					imp = fmt.Sprintf("%s %s", pkgname, imp) // 文件夹名与包名不同
				}
				f.imports[imp] = true

				if i == 0 {
					e.req = t.String()
				} else {
					if t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
						return errors.Errorf("method %s parameter resp isn't struct pointer", m.Name)
					}
					e.resp = t.Elem().String()
				}
			}
		} else {
			return errors.Errorf("method %s has invalid input parameter number", m.Name)
		}

		if m.Type.NumOut() == 2 {
			if m.Type.Out(0).String() != "int" {
				return errors.Errorf("method %s first output parameter isn't int", m.Name)
			} else if !implemented[error](m.Type.Out(1)) {
				return errors.Errorf("method %s second output parameter isn't error", m.Name)
			}
		} else {
			return errors.Errorf("method %s has invalid output parameter number", m.Name)
		}

		f.methods = append(f.methods, e)
	}

	return b.Build()
}

var prefix2httpmethod = map[string]string{
	"Get":  "GET",
	"Post": "POST",
}

func fnname(origin string) (name, kind string, err error) {
	for p, m := range prefix2httpmethod {
		if strings.HasPrefix(origin, p) {
			name = strings.TrimPrefix(origin, p)
			for name[0] == '_' {
				name = name[1:]
			}
			if !isuper(name[0]) {
				return "", "", errors.Errorf("invalid method name %s", origin)
			}
			return name, m, nil
		}
	}
	return "", "", nil
}

func checkParaType(t reflect.Type) error {
	switch k := t.Kind(); k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
	case reflect.Array:
	case reflect.Chan:
		return errors.Errorf("not support chan type")
	case reflect.Func:
		return errors.Errorf("not support chan type")
	case reflect.Interface:
	case reflect.Map:
	case reflect.Pointer:
		return checkParaType(t.Elem())
	case reflect.Slice:
	case reflect.String:
	case reflect.Struct:
		n := t.NumField()
		for i := range n {
			if f := t.Field(i); f.IsExported() {
				if err := checkParaType(f.Type); err != nil {
					return err
				}
			}
		}
	case reflect.UnsafePointer:
	default:
		return errors.Errorf("not support %s type", k.String())
	}
	if t.Name() == "" {
		return errors.Errorf("not suport anonymous type parameter")
	}
	return nil
}

// isuper 大写字符
func isuper[C byte | rune](c C) bool {
	if 'a' <= c && c <= 'z' {
		return false
	} else if 'A' <= c && c <= 'Z' {
		return true
	} else {
		panic(fmt.Sprintf("not support char %d", c))
	}
}

// touper 转为大写字符
func touper[C byte | rune](c C) C {
	if isuper(c) {
		return c
	} else {
		return 'A' + c - 'a'
	}
}

// implemented t实现接口I
func implemented[I any](t reflect.Type) bool {
	dst := reflect.TypeOf((*I)(nil)).Elem()
	if kind := dst.Kind(); kind != reflect.Interface {
		panic(fmt.Sprintf("type %s not interface", kind.String()))
	}
	return t.Implements(dst)
}

// pkginfo 获取包信息
func pkginfo(t reflect.Type) (pkgpath, pkgname string, err error) {
	if t == nil {
		type a struct{}
		return pkginfo(reflect.TypeOf(a{}))
	} else {
		if t.Kind() == reflect.Pointer {
			return pkginfo(t.Elem())
		}

		if t.Name() == "" {
			return "", "", errors.Errorf("type %s isn't declared", t.String())
		}
		pkgname = strings.TrimSuffix(t.String(), "."+t.Name())
		pkgpath = t.PkgPath()
		return pkgpath, pkgname, nil
	}
}

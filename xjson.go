// Package xjson mainly provides a way to load a "JSON-directory".
// It some cases, it may be convenient to have JSON data spread
// over multiple files. [Read] allows to read such a directory.
// The filesystem tree is converted to a JSON object: directories
// are `map[string]any`; files are `any`.
//
// For example,
//
//   $ echo '"foo"' > foo/bar/baz.json
//
// would yield:
//
//   {"foo" : { "bar" : { "baz" : "foo" } } }
//
// Directories are parsed essentially breadth-first: [Read] first
// loads all JSON files located directly under this directory,
// then move on to load the sub-directories. It means the content of
// a directory `input/foo/` can overload the content of a file
// `input/foo.json`.
//
// TODO more tests, documentation, README.md, etc.
// also, can we get rid of the special case for root?
package xjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var Ext = ".json"

func splitPath(path string) []string {
	return strings.Split(path, string(os.PathSeparator))
}

func TrimExt(fn string) string {
	return strings.TrimSuffix(fn, filepath.Ext(fn))
}

func isRoot(ind, fn string) bool {
	return filepath.Clean(ind) == filepath.Clean(TrimExt(fn))
}

var ErrBadPath = errors.New("bad path")
var ErrBadType = errors.New("bad type")

// maybe type Path []string and make this
// a .String(). and/or handle '.' in xs[i], etc.
func PathString(xs []string) string {
	return strings.Join(xs, ".")
}

// Deep get mechanism
func Get[T any](db map[string]any, xs []string) (T, error) {
	var p map[string]any

	// "zero" value for type T
	tnil := *new(T)

	p = db
	for n, x := range xs {
		q, ok := p[x]
		if !ok {
			return tnil, fmt.Errorf(
				"%w: %s", ErrBadPath, PathString(xs[:n+1]),
			)
		}

		if n == len(xs)-1 {
			r, ok := q.(T)
			if !ok {
				return tnil, fmt.Errorf(
					"%w: '%s'; got '%T', expected '%T'", ErrBadType,
					PathString(xs[:n+1]), p[x], r,
				)
			}
			return r, nil
		}

		r, ok := q.(map[string]any)
		if !ok {
			return tnil, fmt.Errorf(
				"%w: %s", ErrBadPath, PathString(xs[:n+1]),
			)
		}
		p = r
	}

	return tnil, fmt.Errorf(
		"%w: %s", ErrBadPath, PathString(xs),
	)
}

// uint8 may have been sufficient already
type SetFlags uint16
const (
	// when v is a map[string]any, and the leaf pointed to by xs is
	// map[string]any as well, (shallow) merge the two maps.
	//
	// TODO: maybe we'll want a "deep" merge flag?
	MergeMaps = 1 << iota

	// when v is a []T and the leaf pointed to by xs is
	// a []T as well, append v to the leaf
	AppendArrays

	// if while moving through xs we stumble on something which
	// is not a map[string]any (e.g. a string, or an array), remove
	// this value
	ForceThrough
)

// TODO: tests
func SetF[T any](db map[string]any, xs[]string, v any, flags SetFlags) error {
	var p map[string]any

	p = db
	for n, x := range xs {
		if n == len(xs)-1 {
			if (flags & MergeMaps == MergeMaps) {
				w, ok1 := v.(map[string]any)
				q, ok2 := p[x].(map[string]any)
				if ok1 && ok2 {
					for k, x := range w {
						q[k] = x
					}
					continue
				}
			}
			if (flags & AppendArrays == AppendArrays) {
				w, ok1 := v.([]T)
				q, ok2 := p[x].([]T)
				if ok1 && ok2 {
					p[x] = append(q, w...)
					continue
				}
			}
			p[x] = v
			return nil
		}

		// create next leaf if it doesn't exists
		q, ok := p[x]
		if !ok {
			p[x] = make(map[string]any)
			p = p[x].(map[string]any)
			continue
		}

		// current leaf isn't a a map[string]any,
		// can't move further. Fset() ("force set")
		// would swap that entry with a fresh map[string]any
		// instead.
		r, ok := q.(map[string]any)
		if !ok {
			if flags & ForceThrough != ForceThrough {
				return fmt.Errorf(
					"%w: %s", ErrBadPath, PathString(xs[:n+1]),
				)
			}
			p[x] = make(map[string]any)
			p = p[x].(map[string]any)
			continue
		}
		p = r
	}

	// xs is empty; would make sense to update db, but
	// that's good enough
	return nil
}

// Good defaults
func Set(db map[string]any, xs[]string, v any) error {
	return SetF[any](db, xs, v, MergeMaps)
}

func Store(ind, fn string, y any, db map[string]any) error {
	rel, err := filepath.Rel(ind, fn)
	if err != nil {
		return err
	}

	// special case: root
	if isRoot(ind, fn) {
		z, ok := y.(map[string]any)
		if !ok {
			return fmt.Errorf("root isn't a hash")
		}
		// merge maps
		for k, v := range z {
			db[k] = v
		}
		return nil
	}

	rel = rel
	xs := splitPath(strings.TrimSuffix(rel, filepath.Ext(fn)))
	/*
	xs := splitPath(
		strings.TrimPrefix(ind, strings.TrimSuffix(fn, filepath.Ext(fn))),
	)
	*/

	return Set(db, xs, y)
}

// TODO: test line/column numbers in error.
func ReadFile(fn string) (any, error) {
	bs, err := os.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", fn, err)
	}

	var v any
	err = json.Unmarshal(bs, &v)

	// error − if any − is expected to be a json.SyntaxError,
	// which contains an offset, from which we can compute line
	// and column numbers.
	var jErr *json.SyntaxError
	if errors.As(err, &jErr) {
		off := jErr.Offset
		// https://github.com/golang/go/issues/43513#issuecomment-755754498
		line := 1 + bytes.Count(bs[:off], []byte("\n"))
		col := int64(1) + off - int64(bytes.LastIndex(bs[:off], []byte("\n")) + len("\n"))

		return v, fmt.Errorf("unmarshaling %s:%d:%d %w", fn, line, col, err)
	}

	return v, err
}

func doWriteFile(fn string, xs []byte) error {
	dir := filepath.Dir(fn)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}

	if err := os.WriteFile(fn, xs, 0660); err != nil {
		return fmt.Errorf("writing to %s: %w", fn, err)
	}

	return nil
}

func WriteFile(fn string, v any) error {
	xs, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}

	return doWriteFile(fn, xs)
}

// XXX may parametrize the indent?
func WriteIndentFile(fn string, v any) error {
	xs, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}

	return doWriteFile(fn, xs)
}

func ReadFileT(t *testing.T, fn string) any {
	v, err := ReadFile(fn)
	if err != nil {
		t.Fatalf("cannot read '%s': %s", fn, err.Error())
	}

	return v
}

// ReadFile slurps the file pointed to by fn, and attempts to
// JSON-unmarshal it to "to".
func ReadAndStoreFile(ind, fn string, db map[string]any) error {
	v, err := ReadFile(fn)
	if err != nil {
		return err
	}

	return Store(ind, fn, v, db)
}

func ReadAndStoreDir(ind string, db map[string]any) error {
	err := filepath.Walk(ind, func(fn string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return ReadAndStoreFile(ind, fn, db)
	})

	return err
}

func GetPaths(path string) (dn, fn string) {
	fn, dn = path, path

	// XXX test db/
	if strings.HasSuffix(path, Ext) {
		dn = strings.TrimSuffix(strings.TrimRight(path, "/"), Ext)
	} else {
		fn = path + Ext
	}

	return dn, fn
}

// if `path` is "path/to/db", tries to read, in that order:
//	1. path/to/db.json
//	2. path/to/db/
// both reads may succeed. values from db/ would eventually
// supersed those from db.json.
func DoReadAndStore(path string, db map[string]any) error {
	dn, fn := GetPaths(path)

	// keep going if fn doesn't exist
	err0 := ReadAndStoreFile(dn, fn, db);
	if err0 != nil && !errors.Is(err0, os.ErrNotExist) {
		return err0
	}

	// keep going if dn doesn't exist
	err1 := ReadAndStoreDir(dn, db)
	if err1 != nil && !errors.Is(err1, os.ErrNotExist) {
		return err1
	}

	// fn AND dn do not exist: we expect at least
	// one of them to.
	if err0 != nil && err1 != nil {
		return errors.Join(err0, err1)
	}

	return nil
}

func Read(path string) (map[string]any, error) {
	db := make(map[string]any)

	return db, DoReadAndStore(path, db)
}

package xjson

import (
	"fmt"
	"testing"

	"github.com/mbivert/ftests"
)

// StoreT wraps Store to ease tests
func StoreT(ind, fn string, y any, db map[string]any) (map[string]any, error) {
	err := Store(ind, fn, y, db)
	return db, err
}

// TODO: errors
func TestStore(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "file in current directory, empty DB",
			Fun:      StoreT,
			Args:     []any{".", "foo.json", "bar", make(map[string]any)},
			Expected: []any{
				map[string]any{
					"foo" : "bar",
				},
				nil,
			},
		},
		{
			Name:    "file at the base of input directory, empty DB",
			Fun:      StoreT,
			Args:     []any{"path/to/", "path/to/foo.json", "bar", make(map[string]any)},
			Expected: []any{
				map[string]any{
					"foo" : "bar",
				},
				nil,
			},
		},
		{
			Name:    "file at the base of input directory, empty DB bis",
			Fun:      StoreT,
			Args:     []any{"path/to", "path/to/foo.json", "bar", make(map[string]any)},
			Expected: []any{
				map[string]any{
					"foo" : "bar",
				},
				nil,
			},
		},
		{
			Name:    "file one layer of input directory, empty DB",
			Fun:      StoreT,
			Args:     []any{"path", "path/to/foo.json", "bar", make(map[string]any)},
			Expected: []any{
				map[string]any{
					"to" : map[string]any{
						"foo" : "bar",
					},
				},
				nil,
			},
		},
		{
			Name:    "file one layer of input directory, non-empty DB",
			Fun:      StoreT,
			Args:     []any{
				"path", "path/to/foo.json", "bar",
				map[string]any{
					"funky" : "town",
				},
			},
			Expected: []any{
				map[string]any{
					"to" : map[string]any{
						"foo" : "bar",
					},
					"funky" : "town",
				},
				nil,
			},
		},
		{
			Name:    "file one layer of input directory, non-empty DB bis",
			Fun:      StoreT,
			Args:     []any{
				"path", "path/to/foo.json", "bar",
				map[string]any{
					"to" : map[string]any{
						"funky" : "town",
					},
				},
			},
			Expected: []any{
				map[string]any{
					"to" : map[string]any{
						"foo" : "bar",
						"funky" : "town",
					},
				},
				nil,
			},
		},
		{
			Name:    "file one layer of input directory, non-empty DB, override",
			Fun:      StoreT,
			Args:     []any{
				"path", "path/to/foo.json", "bar",
				map[string]any{
					"to" : map[string]any{
						"foo" : "baz",
					},
				},
			},
			Expected: []any{
				map[string]any{
					"to" : map[string]any{
						"foo" : "bar",
					},
				},
				nil,
			},
		},
		{
			Name:    "map merge",
			Fun:      StoreT,
			Args:     []any{
				"path", "path/to.json",
				map[string]any{
					"bar" : "baz",
				},
				map[string]any{
					"to" : map[string]any{
						"foo" : "bar",
					},
				},
			},
			Expected: []any{
				map[string]any{
					"to" : map[string]any{
						"foo" : "bar",
						"bar" : "baz",
					},
				},
				nil,
			},
		},
		{
			Name:    "special case: root not a hash",
			Fun:      StoreT,
			Args:     []any{"path/to/db/", "path/to/db.json", "bar", make(map[string]any)},
			Expected: []any{
				map[string]any{},
				fmt.Errorf("root isn't a hash"),
			},
		},
		{
			Name:    "special case: root",
			Fun:      StoreT,
			Args:     []any{
				"path/to/db/", "path/to/db.json", map[string]any{"foo": "bar"},
				make(map[string]any),
			},
			Expected: []any{
				map[string]any{"foo":"bar"},
				nil,
			},
		},
	})
}

func TestRead(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "DB0",
			Fun:      Read,
			Args:     []any{"testdata/testDB0/db"},
			Expected: []any{
				ReadFileT(t, "testdata/testDB0/expected.json"),
				nil,
			},
		},
	})
}

func TestGet(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "empty-empty",
			Fun:      Get[any],
			Args:     []any{map[string]any{}, []string{}},
			Expected: []any{
				nil,
				fmt.Errorf("%w: %s", BadPathError, ""),
			},
		},
		{
			Name:    "empty-path",
			Fun:      Get[any],
			Args:     []any{map[string]any{}, []string{"path"}},
			Expected: []any{
				nil,
				fmt.Errorf("%w: %s", BadPathError, "path"),
			},
		},
		{
			Name:    "non-empty-empty",
			Fun:      Get[any],
			Args:     []any{map[string]any{"foo":"bar"}, []string{}},
			Expected: []any{
				nil,
				fmt.Errorf("%w: %s", BadPathError, ""),
			},
		},
		{
			Name:    "valid path wrong type",
			Fun:      Get[int],
			Args:     []any{map[string]any{"foo":"bar"}, []string{"foo"}},
			Expected: []any{
				0,
				fmt.Errorf("%w: %s", BadTypeError, "string"),
			},
		},
		{
			Name:    "valid path valid type",
			Fun:      Get[string],
			Args:     []any{map[string]any{"foo":"bar"}, []string{"foo"}},
			Expected: []any{
				"bar",
				nil,
			},
		},
		{
			Name:    "valid nested path",
			Fun:      Get[string],
			Args:     []any{
				map[string]any{
					"foo": map[string]any{
						"bar" : "baz",
					},
				},
				[]string{"foo", "bar"},
			},
			Expected: []any{
				"baz",
				nil,
			},
		},
		{
			Name:    "cannot go this deep",
			Fun:      Get[string],
			Args:     []any{
				map[string]any{
					"foo": "bar",
				},
				[]string{"foo", "bar"},
			},
			Expected: []any{
				"",
				fmt.Errorf("%w: %s", BadPathError, "foo"),
			},
		},
		{
			Name:    "cannot go this deep bis",
			Fun:      Get[string],
			Args:     []any{
				map[string]any{},
				[]string{"foo", "bar"},
			},
			Expected: []any{
				"",
				fmt.Errorf("%w: %s", BadPathError, "foo"),
			},
		},
	})
}

// SeFtT wraps Set to ease tests
func SetFT[T any](db map[string]any, xs[]string, v any, flags SetFlags) (map[string]any, error) {
	err := SetF[T](db, xs, v, flags)
	return db, err
}

func TestSetF(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "empty path - no flags",
			Fun:      SetFT[any],
			Args:     []any{map[string]any{}, []string{}, "world", SetFlags(0)},
			Expected: []any{
				map[string]any{},
				nil,
			},
		},
		{
			Name:    "AppendArrays",
			Fun:      SetFT[string],
			Args:     []any{
				map[string]any{
					"foo" : []string{"hello"},
				},
				[]string{"foo"},
				[]string{"world"},
				SetFlags(AppendArrays),
			},
			Expected: []any{
				map[string]any{
					"foo": []string{"hello", "world"},
				},
				nil,
			},
		},
	})
}

// SetT wraps Set to ease tests
func SetT(db map[string]any, xs[]string, v any) (map[string]any, error) {
	err := Set(db, xs, v)
	return db, err
}

func TestSet(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "empty path",
			Fun:      SetT,
			Args:     []any{map[string]any{}, []string{}, "world"},
			Expected: []any{
				map[string]any{},
				nil,
			},
		},
		{
			Name:    "single depth path",
			Fun:      SetT,
			Args:     []any{map[string]any{}, []string{"hello"}, "world"},
			Expected: []any{
				map[string]any{"hello": "world"},
				nil,
			},
		},
		{
			Name:    "deep path",
			Fun:      SetT,
			Args:     []any{map[string]any{}, []string{"hello", "corporation"}, "world"},
			Expected: []any{
				map[string]any{
					"hello": map[string]any{"corporation": "world"},
				},
				nil,
			},
		},
	})
}

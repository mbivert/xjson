package xjson

import (
	"fmt"
	"testing"

	"github.com/mbivert/ftests"
)

func Test_GetPaths(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "empty",
			Fun:      GetPaths,
			Args:     []any{""},
			Expected: []any{"", ".json"},
		},
		// :shrug:
		{
			Name:    ".",
			Fun:      GetPaths,
			Args:     []any{"."},
			Expected: []any{".", "..json"},
		},
		{
			Name:    "foo",
			Fun:      GetPaths,
			Args:     []any{"foo"},
			Expected: []any{"foo", "foo.json"},
		},
		{
			Name:    "foo.json",
			Fun:      GetPaths,
			Args:     []any{"foo.json"},
			Expected: []any{"foo", "foo.json"},
		},
		{
			Name:    "bar/foo.json",
			Fun:      GetPaths,
			Args:     []any{"bar/foo.json"},
			Expected: []any{"bar/foo", "bar/foo.json"},
		},
	})
}

// StoreT wraps Store to ease tests
func StoreT(ind, fn string, y any, db map[string]any) (map[string]any, error) {
	err := Store(ind, fn, y, db)
	return db, err
}

// TODO: errors
func TestStore(t *testing.T) {
	ftests.Run(t, []ftests.Test{
		{
			Name:    "file_in_current_directory,_empty_DB",
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
			Name:    "file_at_the_base_of_input_directory,_empty_DB",
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
			Name:    "file_at_the_base_of_input_directory,_empty_DB_bis",
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
			Name:    "file_one_layer_of_input_directory,_empty_DB",
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
			Name:    "file_one_layer_of_input_directory,_non-empty_DB",
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
			Name:    "file_one_layer_of_input_directory,_non-empty_DB_bis",
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
			Name:    "file_one_layer_of_input_directory,_non-empty_DB,_override",
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
			Name:    "map_merge",
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
			Name:    "special_case:_root_not_a_hash",
			Fun:      StoreT,
			Args:     []any{"path/to/db/", "path/to/db.json", "bar", make(map[string]any)},
			Expected: []any{
				map[string]any{},
				fmt.Errorf("root isn't a hash"),
			},
		},
		{
			Name:    "special_case:_root",
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
				fmt.Errorf("%w: %s", ErrBadPath, ""),
			},
		},
		{
			Name:    "empty-path",
			Fun:      Get[any],
			Args:     []any{map[string]any{}, []string{"path"}},
			Expected: []any{
				nil,
				fmt.Errorf("%w: %s", ErrBadPath, "path"),
			},
		},
		{
			Name:    "non-empty-empty",
			Fun:      Get[any],
			Args:     []any{map[string]any{"foo":"bar"}, []string{}},
			Expected: []any{
				nil,
				fmt.Errorf("%w: %s", ErrBadPath, ""),
			},
		},
		{
			Name:    "valid_path_wrong_type",
			Fun:      Get[int],
			Args:     []any{map[string]any{"foo":"bar"}, []string{"foo"}},
			Expected: []any{
				0,
				fmt.Errorf("%w: 'foo'; got 'string', expected 'int'", ErrBadType),
			},
		},
		{
			Name:    "valid_path_valid_type",
			Fun:      Get[string],
			Args:     []any{map[string]any{"foo":"bar"}, []string{"foo"}},
			Expected: []any{
				"bar",
				nil,
			},
		},
		{
			Name:    "valid_nested_path",
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
			Name:    "cannot_go_this_deep",
			Fun:      Get[string],
			Args:     []any{
				map[string]any{
					"foo": "bar",
				},
				[]string{"foo", "bar"},
			},
			Expected: []any{
				"",
				fmt.Errorf("%w: %s", ErrBadPath, "foo"),
			},
		},
		{
			Name:    "cannot_go_this_deep_bis",
			Fun:      Get[string],
			Args:     []any{
				map[string]any{},
				[]string{"foo", "bar"},
			},
			Expected: []any{
				"",
				fmt.Errorf("%w: %s", ErrBadPath, "foo"),
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
			Name:    "empty_path_-_no_flags",
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
			Name:    "empty_path",
			Fun:      SetT,
			Args:     []any{map[string]any{}, []string{}, "world"},
			Expected: []any{
				map[string]any{},
				nil,
			},
		},
		{
			Name:    "single_depth_path",
			Fun:      SetT,
			Args:     []any{map[string]any{}, []string{"hello"}, "world"},
			Expected: []any{
				map[string]any{"hello": "world"},
				nil,
			},
		},
		{
			Name:    "deep_path",
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

func TestReadFile(t *testing.T) {
	// TODO
}

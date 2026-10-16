package spec

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	json "github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSearch = `{
  "search": {
    "url": {"paths": [
      {"path": "/{index}/_search", "methods": ["GET", "POST"], "parts": {"index": {"type": "list"}}},
      {"path": "/_search", "methods": ["GET", "POST"]}
    ]},
    "params": {"q": {"type": "string"}},
    "stability": "stable",
    "visibility": "public"
  }
}`

func TestParseFileValid(t *testing.T) {
	eps, err := ParseFile("whatever.json", []byte(validSearch))
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "search", eps[0].Name)
	assert.Equal(t, "stable", eps[0].Endpoint.Stability)
	assert.Equal(t, "public", eps[0].Endpoint.Visibility)
	assert.Equal(t, []string{"index"}, CollectParts(eps[0].Endpoint.URL.Paths))
	_, ok := eps[0].Endpoint.Params["q"]
	assert.True(t, ok)
}

func TestParseFileMultipleKeysSorted(t *testing.T) {
	raw := `{
	  "zeta": {"url": {"paths": [{"path": "/z", "methods": ["GET"]}]}},
	  "alpha": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}}
	}`
	eps, err := ParseFile("f.json", []byte(raw))
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, "alpha", eps[0].Name)
	assert.Equal(t, "zeta", eps[1].Name)
}

func TestParseFileRenamedFileAccepted(t *testing.T) {
	_, err := ParseFile("totally-unrelated.json", []byte(validSearch))
	require.NoError(t, err)
}

func TestParseFileErrors(t *testing.T) {
	cases := map[string]string{
		"empty object":         `{}`,
		"null root":            `null`,
		"scalar root":          `"x"`,
		"array root":           `[]`,
		"null endpoint":        `{"a": null}`,
		"missing paths":        `{"a": {"params": {}}}`,
		"empty paths":          `{"a": {"url": {"paths": []}}}`,
		"empty path template":  `{"a": {"url": {"paths": [{"path": "", "methods": ["GET"]}]}}}`,
		"missing methods":      `{"a": {"url": {"paths": [{"path": "/a"}]}}}`,
		"empty method":         `{"a": {"url": {"paths": [{"path": "/a", "methods": [""]}]}}}`,
		"empty key":            `{"": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}}}`,
		"placeholder mismatch": `{"a": {"url": {"paths": [{"path": "/{index}", "methods": ["GET"]}]}}}`,
		"parts not object":     `{"a": {"url": {"paths": [{"path": "/{i}", "methods": ["GET"], "parts": {"i": 5}}]}}}`,
		"null part":            `{"a": {"url": {"paths": [{"path": "/{i}", "methods": ["GET"], "parts": {"i": null}}]}}}`,
		"param not object":     `{"a": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}, "params": {"p": 5}}}`,
		"null param":           `{"a": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}, "params": {"p": null}}}`,
		"bad param metadata":   `{"a": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}, "params": {"p": {"type": 5}}}}`,
		"trailing data":        validSearch + `x`,
		"duplicate key":        `{"a": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}}, "a": {"url": {"paths": [{"path": "/b", "methods": ["GET"]}]}}}`,
		"malformed":            `{not json`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseFile("f.json", []byte(raw))
			assert.Error(t, err, "case %q should fail", name)
		})
	}
}

func TestParseFileContradictoryRepeatedPaths(t *testing.T) {
	raw := `{"a": {"url": {"paths": [
	  {"path": "/{i}", "methods": ["GET"], "parts": {"i": {"type": "string"}}},
	  {"path": "/{i}", "methods": ["POST"], "parts": {"i": {"type": "list"}}}
	]}}}`
	_, err := ParseFile("f.json", []byte(raw))
	assert.Error(t, err)
}

func TestParseFileRepeatedCompatiblePaths(t *testing.T) {
	raw := `{"a": {"url": {"paths": [
	  {"path": "/{i}", "methods": ["GET"], "parts": {"i": {"type": "string"}}},
	  {"path": "/{i}", "methods": ["POST"], "parts": {"i": {"type": "string"}}}
	]}}}`
	_, err := ParseFile("f.json", []byte(raw))
	require.NoError(t, err)
}

func TestParseFileUnknownParamTypeLoads(t *testing.T) {
	raw := `{"a": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}, "params": {"p": {"type": "mystery"}}}}`
	eps, err := ParseFile("f.json", []byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "mystery", eps[0].Endpoint.Params["p"].Type)
}

func TestChecked(t *testing.T) {
	allow := map[string]bool{"streams.status": true, "streams.private": true}
	cases := []struct {
		name       string
		visibility string
		want       bool
	}{
		{"public", "public", true},
		{"legacy", "", true},
		{"private", "private", false},
		{"_internal.x", "public", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ep := NamedEndpoint{Name: c.name, Endpoint: Endpoint{Visibility: c.visibility}}
			assert.Equal(t, c.want, Checked(ep, allow))
		})
	}
	assert.False(t, Checked(NamedEndpoint{Name: "_common"}, allow))
	assert.False(t, Checked(NamedEndpoint{Name: "_internal.foo"}, allow))
	assert.True(t, Checked(NamedEndpoint{Name: "streams.status", Endpoint: Endpoint{Visibility: "feature_flag"}}, allow))
	assert.False(t, Checked(NamedEndpoint{Name: "streams.other", Endpoint: Endpoint{Visibility: "feature_flag"}}, allow))
	// allowlist must not override private visibility
	assert.False(t, Checked(NamedEndpoint{Name: "streams.private", Endpoint: Endpoint{Visibility: "private"}}, allow))
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

func TestLoadDirValid(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "search.json", validSearch)
	writeFile(t, dir, AuxiliaryFile, `{"params": {"pretty": {"type": "boolean"}}}`)
	cat, err := LoadDir(dir)
	require.NoError(t, err)
	assert.Equal(t, 1, cat.Files)
	assert.Equal(t, 1, cat.AuxiliaryFiles)
	assert.Contains(t, cat.Endpoints, "search")
}

func TestLoadDirExcludedOnlyLoads(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "priv.json", `{"priv": {"url": {"paths": [{"path": "/p", "methods": ["GET"]}]}, "visibility": "private"}}`)
	cat, err := LoadDir(dir)
	require.NoError(t, err)
	assert.Len(t, cat.Endpoints, 1)
	assert.False(t, Checked(cat.Endpoints["priv"], nil))
}

func TestLoadDirErrors(t *testing.T) {
	t.Run("missing dir", func(t *testing.T) {
		_, err := LoadDir(filepath.Join(t.TempDir(), "nope"))
		assert.Error(t, err)
	})
	t.Run("empty dir", func(t *testing.T) {
		_, err := LoadDir(t.TempDir())
		assert.Error(t, err)
	})
	t.Run("not a dir", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(f, []byte("x"), 0o644))
		_, err := LoadDir(f)
		assert.Error(t, err)
	})
	t.Run("malformed endpoint", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "bad.json", `{not valid`)
		_, err := LoadDir(dir)
		assert.Error(t, err)
	})
	t.Run("bad auxiliary", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "search.json", validSearch)
		writeFile(t, dir, AuxiliaryFile, `{not valid`)
		_, err := LoadDir(dir)
		assert.Error(t, err)
	})
	t.Run("unknown metadata file is parsed as endpoint", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "search.json", validSearch)
		writeFile(t, dir, "metadata.json", `{"version": "9.0.0"}`)
		_, err := LoadDir(dir)
		assert.Error(t, err, "scalar metadata masquerading as an endpoint must fail")
	})
	t.Run("only auxiliary", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, AuxiliaryFile, `{}`)
		_, err := LoadDir(dir)
		assert.Error(t, err)
	})
	t.Run("cross-file duplicate", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "a.json", `{"dup": {"url": {"paths": [{"path": "/a", "methods": ["GET"]}]}}}`)
		writeFile(t, dir, "b.json", `{"dup": {"url": {"paths": [{"path": "/b", "methods": ["GET"]}]}}}`)
		_, err := LoadDir(dir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate endpoint")
	})
}

type fakeDirEntry struct {
	name  string
	mode  fs.FileMode
	isDir bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.isDir }
func (f fakeDirEntry) Type() fs.FileMode          { return f.mode }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, errors.New("not implemented") }

func TestLoadDirInjectedReadError(t *testing.T) {
	_, err := loadDir("anything", func(string) ([]fs.DirEntry, error) {
		return nil, errors.New("boom")
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestLoadDirNonRegularJSON(t *testing.T) {
	_, err := loadDir("anything", func(string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{fakeDirEntry{name: "weird.json", mode: fs.ModeNamedPipe}}, nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a regular file")
}

func TestCollectPartsSortedUnique(t *testing.T) {
	paths := []Path{
		{Path: "/{b}/{a}", Parts: map[string]json.RawMessage{"b": nil, "a": nil}},
		{Path: "/{a}", Parts: map[string]json.RawMessage{"a": nil}},
	}
	assert.Equal(t, []string{"a", "b"}, CollectParts(paths))
	assert.Empty(t, CollectParts(nil))
}

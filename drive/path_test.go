package drive

import (
	"testing"
)

func TestParseDrivePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"/", "", false},
		{"", "", false},
		{".", "", false},
		{"/path/to/file", "path/to/file", false},
		{"path/to/file/", "path/to/file", false},
		{"//multiple//slashes//", "multiple/slashes", false},
		{"/path/with/../relative", "path/relative", false}, // .. should resolve correctly
		{"/path/with/*/asterisk", "", true},
		{"/path/with/?/question", "", true},
	}

	for _, test := range tests {
		result, err := parseDrivePath(test.input)

		if test.wantErr && err == nil {
			t.Errorf("parseDrivePath(%q) expected error but got nil", test.input)
		}

		if !test.wantErr && err != nil {
			t.Errorf("parseDrivePath(%q) unexpected error: %v", test.input, err)
		}

		if result != test.expected {
			t.Errorf("parseDrivePath(%q) got %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input        string
		expectedDir  string
		expectedLeaf string
	}{
		{"", "", ""},
		{"/", "", ""},
		{"file.txt", "", "file.txt"},
		{"dir/file.txt", "dir", "file.txt"},
		{"/dir/file.txt", "dir", "file.txt"},
		{"/nested/path/to/file.txt", "nested/path/to", "file.txt"},
	}

	for _, test := range tests {
		dir, leaf := splitPath(test.input)

		if dir != test.expectedDir {
			t.Errorf("splitPath(%q) dir got %q, want %q", test.input, dir, test.expectedDir)
		}

		if leaf != test.expectedLeaf {
			t.Errorf("splitPath(%q) leaf got %q, want %q", test.input, leaf, test.expectedLeaf)
		}
	}
}

func TestIsRootDirectory(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"/", true},
		{".", true},
		{"file.txt", false},
		{"dir/file.txt", false},
	}

	for _, test := range tests {
		result := isRootDirectory(test.input)

		if result != test.expected {
			t.Errorf("isRootDirectory(%q) got %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestGetRelativePath(t *testing.T) {
	tests := []struct {
		root     string
		path     string
		expected string
	}{
		{"", "file.txt", "file.txt"},
		{"/", "/file.txt", "file.txt"},
		{"root", "root/file.txt", "file.txt"},
		{"dir", "other/file.txt", "other/file.txt"},
		{"root/dir", "root/dir/file.txt", "file.txt"},
		{"root/dir", "root/dir", ""},
	}

	for _, test := range tests {
		result := getRelativePath(test.root, test.path)

		if result != test.expected {
			t.Errorf("getRelativePath(%q, %q) got %q, want %q", test.root, test.path, result, test.expected)
		}
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		dir      string
		leaf     string
		expected string
	}{
		{"", "file.txt", "file.txt"},
		{"dir", "file.txt", "dir/file.txt"},
		{"nested/path", "file.txt", "nested/path/file.txt"},
	}

	for _, test := range tests {
		result := joinPath(test.dir, test.leaf)

		if result != test.expected {
			t.Errorf("joinPath(%q, %q) got %q, want %q", test.dir, test.leaf, result, test.expected)
		}
	}
}

package pkg

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestSyncLocalToR2KeysAreRelativeToSource(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "src", "a.txt"), "a")
	writeTestFile(t, filepath.Join(root, "src", ".env"), "secret")
	writeTestFile(t, filepath.Join(root, "src", "sub", "b.txt"), "b")

	want := []string{"backup/.env", "backup/a.txt", "backup/sub/b.txt"}

	tests := []struct {
		name   string
		cwd    string
		source string
	}{
		{"bare directory", root, "src"},
		{"dot-slash directory", root, "./src"},
		{"trailing slash", root, "src/"},
		{"absolute path", root, filepath.Join(root, "src")},
		{"current directory", filepath.Join(root, "src"), "."},
		{"parent-relative path", filepath.Join(root, "src", "sub"), "../../src"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(tc.cwd)
			fake, bucket := newFakeS3(t, "test-bucket")

			bucket.SyncLocalToR2WithPrefix(tc.source, "backup")

			if got := fake.keys(); !reflect.DeepEqual(got, want) {
				t.Fatalf("uploaded keys = %v, want %v", got, want)
			}
		})
	}
}

func TestSyncR2ToLocalTreatsPrefixAsDirectory(t *testing.T) {
	for _, prefix := range []string{"photos", "photos/"} {
		t.Run(prefix, func(t *testing.T) {
			fake, bucket := newFakeS3(t, "test-bucket")
			fake.put("photos/", nil) // folder marker, as created by the R2 dashboard
			fake.put("photos/a.jpg", []byte("a"))
			fake.put("photos/2024/", nil)
			fake.put("photos/2024/b.jpg", []byte("b"))
			fake.put("photos-old/c.jpg", []byte("c"))
			fake.put("photosets/d.jpg", []byte("d"))

			dest := t.TempDir()
			bucket.SyncR2ToLocalWithPrefix(dest, prefix)

			want := []string{"2024/b.jpg", "a.jpg"}
			if got := listLocalFiles(t, dest); !reflect.DeepEqual(got, want) {
				t.Fatalf("downloaded files = %v, want %v", got, want)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create dir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// listLocalFiles returns the slash-separated paths of all regular files under root.
func listLocalFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(files)
	return files
}

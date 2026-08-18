package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, dir, dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"../secret", "foo/../../etc/passwd", ".."} {
		if _, err := w.Resolve(p); err == nil {
			t.Errorf("expected reject for %q", p)
		}
	}
}

func TestResolveAllowsNestedFile(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, dir, dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write("firmware/main.c", "int main(){}"); err != nil {
		t.Fatal(err)
	}
	got, err := w.Resolve("firmware/main.c")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "firmware", "main.c")
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := New(dir, dir, dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Resolve("escape/secret.txt"); err == nil {
		t.Fatal("symlink escape should be rejected")
	}
}

func TestLoadTemplateIfEmpty(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, dir, dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.Initialized() {
		t.Fatal("empty workspace should not be initialized")
	}
	if err := w.Write("firmware/main.c", "student"); err != nil {
		t.Fatal(err)
	}
	if err := w.LoadTemplateIfEmpty("hello-gpu"); err != nil {
		t.Fatal(err)
	}
	got, err := w.Read("firmware/main.c")
	if err != nil {
		t.Fatal(err)
	}
	if got != "student" {
		t.Fatalf("startup load wiped student file: %q", got)
	}
}

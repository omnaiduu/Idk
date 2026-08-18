package workspace

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileEntry describes a file or directory in the workspace.
type FileEntry struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir,omitempty"`
}

// Workspace is a path-jailed working directory for user sources and build artifacts.
type Workspace struct {
	root      string
	templates string
	repoHDL   string
	repoFW    string
}

// New creates a workspace rooted at root with template and repo asset directories.
func New(root, templatesDir, repoHDL, repoFW string) (*Workspace, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &Workspace{
		root:      root,
		templates: templatesDir,
		repoHDL:   repoHDL,
		repoFW:    repoFW,
	}, nil
}

func (w *Workspace) Root() string        { return w.root }
func (w *Workspace) BuildDir() string    { return filepath.Join(w.root, "build") }
func (w *Workspace) DumpsDir() string    { return filepath.Join(w.root, "dumps") }
func (w *Workspace) HDLDir() string      { return filepath.Join(w.root, "hdl") }
func (w *Workspace) FirmwareDir() string { return filepath.Join(w.root, "firmware") }

func insideJail(root, abs string) bool {
	if abs == root {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(abs+sep, root+sep)
}

// Resolve validates rel and returns an absolute path inside the workspace jail.
func (w *Workspace) Resolve(rel string) (string, error) {
	rel = strings.TrimPrefix(filepath.ToSlash(rel), "/")
	if rel == "" || rel == "." {
		return w.root, nil
	}
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	root, err := filepath.Abs(w.root)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(clean)))
	if err != nil {
		return "", err
	}
	if !insideJail(root, abs) {
		return "", fmt.Errorf("path outside workspace")
	}

	// If the path (or its parent) exists, resolve symlinks and re-check the jail.
	check := abs
	if _, err := os.Lstat(abs); err != nil {
		check = filepath.Dir(abs)
	}
	if resolved, err := filepath.EvalSymlinks(check); err == nil {
		rootResolved := root
		if rr, err := filepath.EvalSymlinks(root); err == nil {
			rootResolved = rr
		}
		final := resolved
		if check == filepath.Dir(abs) {
			final = filepath.Join(resolved, filepath.Base(abs))
		}
		if !insideJail(rootResolved, final) && !insideJail(root, final) {
			return "", fmt.Errorf("path outside workspace")
		}
	}
	return abs, nil
}

// List returns all files under the workspace (relative paths).
func (w *Workspace) List() ([]FileEntry, error) {
	var entries []FileEntry
	err := filepath.WalkDir(w.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == w.root {
			return nil
		}
		rel, err := filepath.Rel(w.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "build/") || strings.HasPrefix(rel, "dumps/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		entries = append(entries, FileEntry{
			Path:  rel,
			Name:  filepath.Base(path),
			IsDir: d.IsDir(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

// Read returns the text content of a jailed file.
func (w *Workspace) Read(rel string) (string, error) {
	abs, err := w.Resolve(rel)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("is a directory")
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Write stores text content at a jailed relative path.
func (w *Workspace) Write(rel, content string) error {
	abs, err := w.Resolve(rel)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// LoadTemplate copies the named template into the workspace.
func (w *Workspace) LoadTemplate(name string) error {
	if strings.Contains(name, "..") || strings.Contains(name, "/") {
		return fmt.Errorf("invalid template name")
	}
	src := filepath.Join(w.templates, name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("template %q not found: %w", name, err)
	}

	dirs := []string{"hdl", "firmware", "dumps", "build"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(w.root, d), 0o755); err != nil {
			return err
		}
	}

	if err := copyTree(src, w.root); err != nil {
		return err
	}
	if err := copyTree(w.repoHDL, filepath.Join(w.root, "hdl")); err != nil {
		return err
	}
	if err := copyTree(w.repoFW, filepath.Join(w.root, "firmware")); err != nil {
		return err
	}
	// Template main.c may live at template root; ensure firmware/main.c exists.
	tplMain := filepath.Join(src, "main.c")
	fwMain := filepath.Join(w.root, "firmware", "main.c")
	if _, err := os.Stat(fwMain); err != nil {
		if _, err2 := os.Stat(tplMain); err2 == nil {
			if err := copyFile(tplMain, fwMain); err != nil {
				return err
			}
		}
	}
	return nil
}

// EnsureDirs creates build and dump output directories.
func (w *Workspace) EnsureDirs() error {
	for _, d := range []string{w.BuildDir(), w.DumpsDir(), w.HDLDir(), w.FirmwareDir()} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Initialized reports whether the workspace already has student sources.
func (w *Workspace) Initialized() bool {
	_, err := os.Stat(filepath.Join(w.root, "firmware", "main.c"))
	return err == nil
}

// LoadTemplateIfEmpty copies hello-gpu on first boot without wiping edits on restart.
func (w *Workspace) LoadTemplateIfEmpty(name string) error {
	if w.Initialized() {
		return nil
	}
	return w.LoadTemplate(name)
}

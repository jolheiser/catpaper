package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog/log"
)

func download(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		log.Info().Msg("initializing cache, subsequent runs should be faster...")
		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			URL:      "https://github.com/catppuccin/wallpapers.git",
			Progress: os.Stderr,
		})
		if err != nil {
			return err
		}
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}

	tree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = tree.Pull(&git.PullOptions{
		Progress: os.Stderr,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	return nil
}

type wallpaper struct {
	Dir    string
	Name   string
	height int
	width  int
}

func (w wallpaper) Path() string {
	return path.Join(w.Dir, w.Name)
}

func (w wallpaper) div() int {
	num := w.height
	if num < w.width {
		num = w.width
	}
	return num / 300
}

func (w wallpaper) Height() int {
	return w.height / w.div()
}

func (w wallpaper) Width() int {
	return w.width / w.div()
}

func newWallpaper(dir, subdir, name string) (wallpaper, error) {
	w := wallpaper{
		Dir:  subdir,
		Name: name,
	}
	fi, err := os.Open(filepath.Join(dir, subdir, name))
	if err != nil {
		return w, err
	}
	defer fi.Close()

	info, _, err := image.DecodeConfig(fi)
	if err != nil {
		return w, fmt.Errorf("could not decode %q: %w", name, err)
	}
	w.height = info.Height
	w.width = info.Width
	return w, nil
}

func wallpapers(dir string) (map[string][]wallpaper, error) {
	wallpaperMap := make(map[string][]wallpaper)

	if err := filepath.WalkDir(dir, func(walkPath string, walkInfo fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if walkInfo.IsDir() {
			if strings.HasPrefix(walkInfo.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasPrefix(walkInfo.Name(), ".") {
			return nil
		}

		walkDir := strings.TrimLeft(strings.TrimPrefix(filepath.Dir(walkPath), dir), "/")
		wp, err := newWallpaper(dir, walkDir, walkInfo.Name())
		if err != nil {
			if errors.Is(err, image.ErrFormat) {
				log.Err(err).Msg("format not registered")
				return nil
			}
			return err
		}

		if wallpaperMap[walkDir] == nil {
			wallpaperMap[walkDir] = make([]wallpaper, 0)
		}
		wallpaperMap[walkDir] = append(wallpaperMap[walkDir], wp)

		return nil
	}); err != nil {
		return nil, err
	}

	return wallpaperMap, nil
}

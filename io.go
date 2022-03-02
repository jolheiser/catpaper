package main

import (
	"errors"
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
	fi, err := os.Open(path.Join(dir, subdir, name))
	if err != nil {
		return w, err
	}
	defer fi.Close()

	info, _, err := image.DecodeConfig(fi)
	if err != nil {
		return w, err
	}
	w.height = info.Height
	w.width = info.Width
	return w, nil
}

func wallpapers(dir string) (map[string][]wallpaper, error) {
	wallpapers := make(map[string][]wallpaper)
	top, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, t := range top {
		if !t.IsDir() || strings.HasPrefix(t.Name(), ".") {
			continue
		}
		subdir := filepath.Join(dir, t.Name())
		names, err := os.ReadDir(subdir)
		if err != nil {
			return nil, err
		}

		wallpapers[t.Name()] = make([]wallpaper, 0, len(wallpapers))
		for _, file := range names {
			wp, err := newWallpaper(dir, t.Name(), file.Name())
			if err != nil {
				return nil, err
			}
			wallpapers[t.Name()] = append(wallpapers[t.Name()], wp)
		}
	}

	return wallpapers, nil
}

package main

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

import flag "github.com/spf13/pflag"

func inputFromArgs(f io.ReadCloser) <-chan *Info {
	files := make(chan *Info)
	go func() {
		for _, arg := range flag.Args() {
			files <- &Info{file: arg}
		}
		close(files)
	}()
	return files
}

// Used by the -r option
func inputFromDir(f io.ReadCloser) <-chan *Info {
	files := make(chan *Info)
	go func() {
		for _, arg := range flag.Args() {
			filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					logger.Print(err)
					if !os.IsPermission(err) {
						return err
					}
				} else if opts.symlinks && d.Type()&fs.ModeType == fs.ModeSymlink || d.Type().IsRegular() {
					files <- &Info{file: path}
				}
				return nil
			})
		}
		close(files)
	}()
	return files
}

// Used by the -i option
func inputFromFile(f io.ReadCloser) <-chan *Info {
	var err error

	if f == nil {
		f = os.Stdin
		if opts.input != "" {
			if f, err = os.Open(opts.input); err != nil {
				logger.Fatal(err)
			}
		}
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	files := make(chan *Info)
	go func() {
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				logger.Fatal(err)
			}
			file := scanner.Text()
			if file != "" {
				files <- &Info{file: scanner.Text()}
			}
		}
		close(files)
		f.Close()
	}()
	return files
}

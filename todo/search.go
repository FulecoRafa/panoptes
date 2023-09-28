package todo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"slices"
)

func getTodos(ctx context.Context) ([]Todo, error) {
    ctx, done := context.WithCancel(ctx)
    defer done()

    /* PIPELINE */
    // Search all files
    paths, errc := walkdirGenerator(ctx)
    for path := range paths {
        fmt.Println(path)
    }
    if err := <- errc; err != nil {
        return nil, err
    }
    return nil, nil
}

func walkdirGenerator(ctx context.Context) (<-chan string, <-chan error) {
    log.Default().Println("Walking dir")
    paths := make(chan string)
    errc := make(chan error, 1)
    root := ctx.Value("root").(string)
    extensions := ctx.Value("fileExtensions").([]string)
    ignoredDirs := ctx.Value("ignoredDirs").([]string)
    go func() {
        defer close(paths)
        errc <- filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
            log.Default().Println("Searching ", path)
            if err != nil {
                return err
            }
            if !info.Mode().IsRegular() {
                return nil
            }
            // Skip directories
            if info.IsDir() && slices.Contains(ignoredDirs, info.Name()) {
                log.Default().Println("Skipping ", path)
                return filepath.SkipDir
            }
            // Skip files whose extensions are not included
            ext := strings.TrimSpace(filepath.Ext(path))
            if !slices.Contains(extensions, ext) {
                log.Default().Println("Skipping ", path)
                return nil
            }
            select {
            case paths <- path:
                log.Default().Printf("Sent path `%s`", path)
            case <- ctx.Done():
                return errors.New("Walk task was cancelled before finishing")
            }
            return nil
        })
    }()
    return paths, errc
}

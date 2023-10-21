package todo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func getTodos(ctx context.Context) ([]Todo, error) {
    ctx, cancel := context.WithCancelCause(ctx)
    defer cancel(nil)

    /* PIPELINE */
    // Search all files
    pathsChan := walkdirGenerator(ctx, cancel)
    for path := range pathsChan {
        fmt.Println(path)
    }
    return nil, nil
}

func parseFile(ctx context.Context, files <-chan sourceFile, results chan<- *sitter.Tree) error {
    for file := range files {
        fileReader, err := os.Open(file.path)
        if err != nil {
            return err
        }

        lang, ok := decideLanguage(file.info)
        if !ok {
            return errors.New("File of unknown language")
        }

        source, err := io.ReadAll(fileReader)
        if err != nil {
            return err
        }

        parser := sitter.NewParser()
        parser.SetLanguage(lang)
        tree, err := parser.ParseCtx(ctx, nil, source)
        if err != nil {
            return err
        }

        select {
        case results <- tree:
        case <- ctx.Done():
            return nil
        }
    }
    return nil
}

func walkdirGenerator(ctx context.Context, cancelCauseFunc context.CancelCauseFunc) (<-chan sourceFile) {
    log.Default().Println("Walking dir")
    pathsChan := make(chan sourceFile)
    root := ctx.Value("root").(string)
    extensions := ctx.Value("fileExtensions").([]string)
    ignoredDirs := ctx.Value("ignoredDirs").([]string)
    go func() {
        defer close(pathsChan)
        err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
            log.Default().Println("Searching ", path)
            if err != nil {
                return err
            }
            if !info.Type().IsRegular() {
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
            case pathsChan <- sourceFile{path: path, info: info}:
                log.Default().Printf("Sent path `%s`", path)
            case <- ctx.Done():
                return errors.New("walk task was cancelled before finishing")
            }
            return nil
        })
        if err != nil  {
            cancelCauseFunc(err)
        }
    }()
    return pathsChan
}

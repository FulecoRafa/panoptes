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
    ctx, done := context.WithCancel(ctx)
    defer done()

    /* PIPELINE */
    // Search all files
    sourceFiles, errc := walkdirGenerator(ctx)
    throughput := ctx.Value("throughtput").(int)
    trees := make(chan *sitter.Tree, throughput)
    wg := sync.WaitGroup{}
    wg.Add(throughput)
    for i := 0; i < throughput; i++ {
        go func() {
            err := parseFile(ctx, sourceFiles, trees)
            if err != nil {
                errc <- err
                done()
            }
            wg.Done()
        }()
    }
    go func() {
        wg.Wait()
        close(trees)
    }()

    for tree := range trees {
        fmt.Println(tree)
    }

    if err := <- errc; err != nil {
        return nil, err
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

func walkdirGenerator(ctx context.Context) (<-chan sourceFile, chan error) {
    log.Default().Println("Walking dir")
    paths := make(chan sourceFile)
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
            case paths <- sourceFile{path: path, info: info}:
                log.Default().Printf("Sent path `%s`", path)
            case <- ctx.Done():
                return errors.New("Walk task was cancelled before finishing")
            }
            return nil
        })
    }()
    return paths, errc
}

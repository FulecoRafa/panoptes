package todo

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fulecorafa/panoptes/pipeline"
	sitter "github.com/smacker/go-tree-sitter"
)

func getTodos(ctx context.Context) ([]Todo, error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

    /* Context variables */
	workerN := ctx.Value("throughput").(int)

	/* PIPELINE */
	// Search all files
	pathsChan := walkdirGenerator(ctx, cancel)
	parserChan := pipeline.RunInParallel(ctx, cancel, workerN, pathsChan, parseFile)
	return nil, nil
}

func parseFile(ctx context.Context, file sourceFile) (*sitter.Tree, error) {
	fileReader, err := os.Open(file.path)
	if err != nil {
		return nil, err
	}

	lang, ok := decideLanguage(file.info)
	if !ok {
		return nil, errors.New("could not detect file language")
	}

	source, err := io.ReadAll(fileReader)
	if err != nil {
		return nil, err
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
    return tree, nil
}

func walkdirGenerator(ctx context.Context, cancelCauseFunc context.CancelCauseFunc) <-chan sourceFile {
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
			case <-ctx.Done():
				return errors.New("walk task was cancelled before finishing")
			}
			return nil
		})
		if err != nil {
			cancelCauseFunc(err)
		}
	}()
	return pathsChan
}

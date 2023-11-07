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

	// Create file parsers
	parserChan := pipeline.RunInParallel(ctx, cancel, workerN, pathsChan, parseFile)

	// Search todo comments in each file
	todoChan := pipeline.RunInParallel(ctx, cancel, workerN, parserChan, getTodoComments)
	// Collect all Todos
	flatChan := pipeline.Flatten(ctx, todoChan)
	todos, err := pipeline.CollectToSlice[[]Todo](ctx, flatChan)
	if err != nil {
		return nil, err
	}
	/* END PIPELINE */

	for _, todo := range todos {
		log.Default().Println(todo)
	}

	return nil, nil
}

func getTodoComments(ctx context.Context, parser *completeParser) ([]Todo, error) {
	// This is a query string for treesitter to...
	// Search all comments
	// Store found comments on `@comment` variable
	// Conditional match `@comment` to regex
	// Regex filters all comments that start like // TODO
	// Whitespaces are ignored
	todoPattern := `
	((comment) @comment
	(#match? @comment "^[\r\n\t\f\v ]*//[\r\n\t\f\v ]*TODO"))
	`
	query, err := sitter.NewQuery([]byte(todoPattern), parser.lang)
	if err != nil {
		return nil, err
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(query, parser.parser.RootNode())
	todos := []Todo{}
	for {
		matcher, found := qc.NextMatch()
		if !found {
			break
		}

		matcher = qc.FilterPredicates(matcher, parser.fileContent)
		for _, capture := range matcher.Captures {
			todos = append(todos, Todo{
				filePath:   parser.path,
				content:    capture.Node.Content(parser.fileContent),
				start:      capture.Node.StartByte(),
				end:        capture.Node.EndByte(),
				startPoint: capture.Node.StartPoint(),
				endPoint:   capture.Node.EndPoint(),
			})
		}
	}
	return todos, nil
}

func parseFile(ctx context.Context, file sourceFile) (*completeParser, error) {
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
	return &completeParser{
		sourceFile:  file,
		fileContent: source,
		parser:      tree,
		lang:        lang,
	}, nil
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

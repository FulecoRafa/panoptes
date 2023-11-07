package todo

import (
	"io/fs"

	sitter "github.com/smacker/go-tree-sitter"
)

type Todo struct {
	filePath   string
	content    string
	start      uint32
	end        uint32
	startPoint sitter.Point
	endPoint   sitter.Point
}

type completeParser struct {
	sourceFile
	fileContent []byte
	parser      *sitter.Tree
	lang        *sitter.Language
}

type sourceFile struct {
	path string
	info fs.DirEntry
}

package todo

import (
	"io/fs"
)

type Todo struct {
	filePath string
	content  string
	line     int
	column   int
}

type sourceFile struct {
	path string
	info fs.DirEntry
}

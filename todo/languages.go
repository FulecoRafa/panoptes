package todo

import (
	"io/fs"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	enry "github.com/src-d/enry/v2"
)

var languages = map[string]*sitter.Language{
    "Go": golang.GetLanguage(),
    "C": c.GetLanguage(),
    "C++": cpp.GetLanguage(),
    "CSS": css.GetLanguage(),
    "Lua": lua.GetLanguage(),
    "HTML":html.GetLanguage(),
    "Shell": bash.GetLanguage(),
    "Java": java.GetLanguage(),
    "Rust": rust.GetLanguage(),
    "C#": csharp.GetLanguage(),
    "Python": python.GetLanguage(),
    "JavaScript": javascript.GetLanguage(),
    "TypeScript": typescript.GetLanguage(),
    "TSX": tsx.GetLanguage(),
}

func decideLanguage(info fs.DirEntry) (*sitter.Language, bool) {
    fileLanguage, _ := enry.GetLanguageByExtension(info.Name())
    lang, ok := languages[fileLanguage]
    return lang, ok
}

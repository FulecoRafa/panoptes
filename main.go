package main

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fulecorafa/panoptes/todo"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
    Use: "panoptes",
    Short: "Easily find all TODOs in your project",
    Long: "A blazingly fast terminal application for finding all your TODOs with context",
    Args: cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        ctx := context.WithValue(cmd.Context(), "tags", []string{"TODO"})
        ctx = context.WithValue(ctx, "fileExtensions", []string{".go", ".json", ".rs"})
        ctx = context.WithValue(ctx, "ignoredDirs", []string{".git", "build"})
        ctx = context.WithValue(ctx, "throughput", 50)
        ctx = context.WithValue(ctx, "root", args[0])
        todos, err := todo.GetTodos(ctx)
        if err != nil {
            panic(err)
        }

        log.Default().Printf("%#v\n", todos)

        model := todo.InitialModel(todos)
        program := tea.NewProgram(model)
        if _, err := program.Run(); err != nil {
            panic(err)
        }
    },
}

func main() {
    if err := command.Execute(); err != nil {
        log.Panic(err)
    }
}


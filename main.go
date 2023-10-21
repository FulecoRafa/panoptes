package main

import (
	"context"
	"log"

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
        ctx = context.WithValue(ctx, "throughtput", 50)
        ctx = context.WithValue(ctx, "root", args[0])
        todo.DisplayTodos(ctx)
    },
}

func main() {
    if err := command.Execute(); err != nil {
        log.Panic(err)
    }
}


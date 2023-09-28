package todo

import (
	"context"
	"fmt"
)

func DisplayTodos(ctx context.Context) error {
    todos, err := getTodos(ctx)
    if err != nil {
        return err
    }
    fmt.Println(todos)
    return nil
}

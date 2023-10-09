package main

import (
	"context"
	"fmt"
	"os"

	"github.com/craftslab/copatcher/cmd"
)

func main() {
	if err := cmd.Run(context.Background()); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

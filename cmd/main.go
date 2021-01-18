package main

import (
	"fmt"

	"github.com/srcabl/posts/internal/bootstrap"
	"github.com/srcabl/posts/internal/config"
)

func main() {

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	boot, err := bootstrap.New(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("boot: %+v", boot)
}

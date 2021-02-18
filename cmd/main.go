package main

import (
	"fmt"
	"os"

	"github.com/srcabl/posts/internal/boot"
	"github.com/srcabl/services/pkg/config"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cfg, err := config.NewService(fmt.Sprintf("%s/config.yml", dir))
	if err != nil {
		panic(err)
	}

	strap, err := boot.New(cfg)
	if err != nil {
		panic(err)
	}

	err = strap.Connect()
	if err != nil {
		panic(err)
	}
	defer func() {
		errs := strap.Shutdown()
		if errs != nil {
			msg := "ERRORS ON SHUTDOWN:"
			for e := range errs {
				msg += fmt.Sprintf(" ---- %+v", e)
			}
			panic(msg)
		}
	}()
}

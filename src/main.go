package main

import (
	"log"

	"github.com/cgouz/pgo-mcp/src/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

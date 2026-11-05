package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("VULNEX Scanner Worker daemon v0.1.0-dev initializing...")
	if len(os.Args) > 1 && os.Args[1] == "version" {
		return
	}
}

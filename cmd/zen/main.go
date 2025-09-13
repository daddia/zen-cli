package main

import (
	"os"

	"github.com/jonathandaddia/zen/internal/zencmd"
)

func main() {
	code := zencmd.Main()
	os.Exit(int(code))
}

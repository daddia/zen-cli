package main

import (
	"os"

	"github.com/daddia/zen/internal/zencmd"
)

func main() {
	code := zencmd.Main()
	os.Exit(int(code))
}

package main

import (
	"fmt"
	"github.com/cowlet/moncow/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is MonCow, a language derived from Monkey :)\n",
		user.Username)
	repl.Start(os.Stdin, os.Stdout)
}

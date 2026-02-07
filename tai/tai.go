package main

import (
	"fmt"
	"os"
)

func main() {
	arguments := GetArguments()

	err := ValidateArguments(arguments)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

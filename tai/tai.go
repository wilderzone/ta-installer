package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	arguments := GetArguments()

	err := ValidateArguments(arguments)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	switch arguments.method {
		case "asset":
			asset, err := CreateAsset(arguments.path, arguments.name)
			if err != nil {
				fmt.Println("Failed to create asset file:", err.Error())
				os.Exit(1)
			}
			path := filepath.Join(arguments.path, arguments.name + ".json")
			file, err := os.OpenFile(path, os.O_CREATE, 0777)
			defer file.Close()
			if err != nil {
				fmt.Println("Failed to create asset file:", err.Error())
				os.Exit(1)
			}
			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "\t")
			err = encoder.Encode(asset)
			if err != nil {
				fmt.Println("Failed to encode asset file:", err.Error())
				os.Exit(1)
			}
			plural := "file:"
			if len(asset.Files) != 1 {
				plural = "files:"
			}
			fmt.Println("Created asset containing", len(asset.Files), plural)
			fmt.Println(path)
	}
}

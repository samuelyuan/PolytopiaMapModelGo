package main

import (
	"fmt"
	polytopiamapmodel "github.com/samuelyuan/polytopiamapmodelgo"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <polytopia_file.state>")
		fmt.Println("Example: go run main.go ../../../examples/my_save.state")
		os.Exit(1)
	}

	filename := os.Args[1]
	fmt.Printf("Testing file: %s\n", filename)

	// Enable debug mode for troubleshooting
	polytopiamapmodel.DebugMode = true

	// Try to read the file (catch any panics from log.Fatal)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("FAILED: %v\n", r)
			os.Exit(1)
		}
	}()

	saveOutput, err := polytopiamapmodel.ReadPolytopiaCompressedFile(filename)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		os.Exit(1)
	}

	// Success!
	fmt.Printf("SUCCESS!\n")
	fmt.Printf("   Map Size: %dx%d\n", saveOutput.MapWidth, saveOutput.MapHeight)
	fmt.Printf("   Game Version: %d\n", saveOutput.GameVersion)
	fmt.Printf("   Player Count: %d\n", len(saveOutput.PlayerData))
	fmt.Printf("   Current Turn: %d\n", saveOutput.MaxTurn)
}

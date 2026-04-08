package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("cerebray server starting...")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("listening on :%s\n", port)
}

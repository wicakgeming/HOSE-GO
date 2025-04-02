package main

import (
	"log"
	"backend/routes" // Impor package routes
)

func main() {
	// Menginisialisasi router dengan SetupRouter
	r := routes.SetupRouter()

	// Jalankan server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

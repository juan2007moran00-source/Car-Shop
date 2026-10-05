package main

import (
	"log"
	"net/http"

	"CarShop/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Archivos estáticos (css, js, imágenes)
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	// Rutas públicas
	mux.HandleFunc("GET /{$}", handlers.Inicio)

	log.Println("Servidor en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

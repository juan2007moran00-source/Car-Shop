package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func Inicio(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/templates/inicio.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	datos := map[string]any{
		"Titulo": "CarShop",
	}
	t.Execute(w, datos)
}

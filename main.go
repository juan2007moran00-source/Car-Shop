package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Mi página en Go</title>
  <style>
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; font-family: system-ui, sans-serif; background: #f3f6fb; color: #172033; }
    main { max-width: 560px; margin: 24px; padding: 40px; border-radius: 18px; background: white; box-shadow: 0 12px 40px #15264a18; }
    h1 { margin-top: 0; color: #1769aa; }
    a { color: #1769aa; }
  </style>
</head>
<body>
  <main>
    <h1>¡Hola desde Go!</h1>
    <p>Esta página básica funciona con un servidor HTTP escrito en Go.</p>
    <p>Ya puedes empezar a personalizarla.</p>
  </main>
</body>
</html>`)
	})

	addr := ":8080"
	log.Printf("Servidor listo en http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

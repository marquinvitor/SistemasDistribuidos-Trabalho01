package main

import (
	"log"
	"net/http"

	"sgrh/internal/handlers" 
	"sgrh/internal/models"   
)

func main() {
	
	db := make(map[string]*models.Departamento)

	
	db["TI"] = &models.Departamento{Nome: "TI"}
	db["RH"] = &models.Departamento{Nome: "RH"}

	
	handler := handlers.NewDepartamentoHandler(db)


	http.HandleFunc("GET /departamentos/{nome}/colaboradores", handler.ListarColab)
	http.HandleFunc("POST /departamentos/{nome}/colaboradores", handler.AddColaborador)
	http.HandleFunc("DELETE /departamentos/{nome}/colaboradores/{id}", handler.DemitirColaborador)
	http.HandleFunc("GET /departamentos/{nome}/folha-salarial", handler.CalcularFolhaSalarial)

	log.Println("Servidor iniciado na porta :8080")
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
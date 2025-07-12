package handlers

import(
	"encoding/json"
	"fmt"
	"sgrh/internal/models"
	"net/http"
	"strconv"
	"log"
)

type DepartamentoHandler struct{
	DB map[string]*models.Departamento
}

func NewDepartamentoHandler(db map[string]*models.Departamento) *DepartamentoHandler{
	return &DepartamentoHandler{DB: db}
}

func(h *DepartamentoHandler) AddColaborador(w http.ResponseWriter, r *http.Request){
	nomeDpto := r.PathValue("nome")
	depto, ok := h.DB[nomeDpto]

	if !ok {
		http.Error(w, "Departamento nao encontrado", http.StatusNotFound)
		return
	}

	var requestBody struct {
		Tipo             string  `json:"tipo"` // "efetivo", "autonomo", "estagiario"
		Id               int     `json:"id"`
		Nome             string  `json:"nome"`
		SalarioMensal    float64 `json:"salario_mensal,omitempty"`
		HorasTrabalhadas int     `json:"horas_trabalhadas,omitempty"`
		ValorHora        float64 `json:"valor_hora,omitempty"`
		AuxilioEstagio   float64 `json:"auxilio_estagio,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var novoColaborador models.Colaborador

	switch requestBody.Tipo {
	case "efetivo":
		novoColaborador = &models.Efetivo{
			ColaboradorBase: models.ColaboradorBase{Id: requestBody.Id, Nome: requestBody.Nome},
			SalarioMensal:   requestBody.SalarioMensal,
		}
	case "autonomo":
		novoColaborador = &models.Autonomo{
			ColaboradorBase:  models.ColaboradorBase{Id: requestBody.Id, Nome: requestBody.Nome},
			HorasTrabalhadas: requestBody.HorasTrabalhadas,
			ValorHora:        requestBody.ValorHora,
		}
	case "estagiario":
		novoColaborador = &models.Estagiario{
			ColaboradorBase: models.ColaboradorBase{Id: requestBody.Id, Nome: requestBody.Nome},
			AuxilioEstagio:  requestBody.AuxilioEstagio,
		}
	default:
		http.Error(w, "Tipo de colaborador inválido", http.StatusBadRequest)
		return
	}

	depto.AdicionarColaborador(novoColaborador)
	log.Printf("Colaborador %s adicionado ao depto %s", novoColaborador.GetNome(), depto.Nome)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Colaborador adicionado com sucesso!")
}

func (h *DepartamentoHandler) ListarColab (w http.ResponseWriter, r *http.Request){
	nomeDpto := r.PathValue("nome")
	depto, ok := h.DB[nomeDpto]

	if !ok {
		http.Error(w, "Departamento nao encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(depto.Colaboradores)
}

func (h *DepartamentoHandler) DemitirColaborador(w http.ResponseWriter, r *http.Request) {
	nomeDepto := r.PathValue("nome")
	idStr := r.PathValue("id")

	depto, ok := h.DB[nomeDepto]
	if !ok {
		http.Error(w, "Departamento não encontrado", http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := depto.DemitirColaborador(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("Colaborador ID %d demitido do depto %s", id, depto.Nome)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Colaborador demitido com sucesso!")
}

func (h *DepartamentoHandler) CalcularFolhaSalarial(w http.ResponseWriter, r *http.Request) {
	nomeDepto := r.PathValue("nome")
	depto, ok := h.DB[nomeDepto]
	if !ok {
		http.Error(w, "Departamento não encontrado", http.StatusNotFound)
		return
	}

	total := depto.CalcularFolhaSalarial()

	response := map[string]interface{}{
		"departamento":           depto.Nome,
		"total_folha_salarial": total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
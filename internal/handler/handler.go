package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/balmiza/poc_pm_tech/internal/domain"
	"github.com/balmiza/poc_pm_tech/internal/service"
)

// DispositivoHandler expõe os casos de uso como endpoints HTTP.
type DispositivoHandler struct {
	svc *service.DispositivoService
}

func NewDispositivoHandler(svc *service.DispositivoService) *DispositivoHandler {
	return &DispositivoHandler{svc: svc}
}

// RegisterRoutes registra todos os endpoints no mux fornecido.
// Requer Go 1.22+ para suporte a method-based routing e path parameters.
func (h *DispositivoHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /clientes/{clienteID}/autorizacao", h.consultarAutorizacao)
	mux.HandleFunc("POST /clientes/{clienteID}/dispositivos", h.associarDispositivo)
	mux.HandleFunc("GET /dispositivos/{dispositivoID}", h.buscarDispositivo)
	mux.HandleFunc("PATCH /dispositivos/{dispositivoID}/habilitar", h.habilitarDispositivo)
	mux.HandleFunc("PATCH /dispositivos/{dispositivoID}/cancelar", h.cancelarDispositivo)
}

// --- tipos de resposta ---

type autorizacaoResponse struct {
	ClienteID     string  `json:"cliente_id"`
	Autorizado    bool    `json:"autorizado"`
	DispositivoID *string `json:"dispositivo_id,omitempty"`
}

type dispositivoResponse struct {
	DispositivoID    string     `json:"dispositivo_id"`
	ClienteID        string     `json:"cliente_id"`
	Estado           string     `json:"estado"`
	DataAssociacao   time.Time  `json:"data_associacao"`
	DataHabilitacao  *time.Time `json:"data_habilitacao,omitempty"`
	DataCancelamento *time.Time `json:"data_cancelamento,omitempty"`
}

type erroResponse struct {
	Erro string `json:"erro"`
}

// --- helpers ---

func toDispositivoResponse(d *domain.Dispositivo) dispositivoResponse {
	return dispositivoResponse{
		DispositivoID:    d.DispositivoID,
		ClienteID:        d.ClienteID,
		Estado:           string(d.Estado),
		DataAssociacao:   d.DataAssociacao,
		DataHabilitacao:  d.DataHabilitacao,
		DataCancelamento: d.DataCancelamento,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeErro(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, erroResponse{Erro: msg})
}

func statusParaErro(err error) int {
	switch {
	case errors.Is(err, domain.ErrDispositivoNaoEncontrado):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrClienteJaPossuiDispositivoAtivoPendente),
		errors.Is(err, domain.ErrTransicaoDeEstadoInvalida):
		return http.StatusConflict
	case errors.Is(err, domain.ErrBiometriaReprovada):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// --- handlers ---

// consultarAutorizacao verifica se o cliente possui dispositivo autorizado.
// GET /clientes/{clienteID}/autorizacao
// Sempre retorna 200; o campo "autorizado" indica o resultado.
func (h *DispositivoHandler) consultarAutorizacao(w http.ResponseWriter, r *http.Request) {
	clienteID := r.PathValue("clienteID")
	res, err := h.svc.ConsultarAutorizacao(clienteID)
	if err != nil {
		writeErro(w, http.StatusInternalServerError, err.Error())
		return
	}
	body := autorizacaoResponse{ClienteID: clienteID, Autorizado: res.Autorizado}
	if res.Autorizado {
		body.DispositivoID = &res.DispositivoID
	}
	writeJSON(w, http.StatusOK, body)
}

// associarDispositivo cria um novo dispositivo em pendente_confirmacao para o cliente.
// POST /clientes/{clienteID}/dispositivos
// 201 em sucesso; 409 se o cliente já tiver dispositivo ativo ou pendente.
func (h *DispositivoHandler) associarDispositivo(w http.ResponseWriter, r *http.Request) {
	clienteID := r.PathValue("clienteID")
	d, err := h.svc.AssociarDispositivo(clienteID)
	if err != nil {
		writeErro(w, statusParaErro(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, toDispositivoResponse(d))
}

// buscarDispositivo retorna o estado atual de um dispositivo.
// GET /dispositivos/{dispositivoID}
func (h *DispositivoHandler) buscarDispositivo(w http.ResponseWriter, r *http.Request) {
	dispositivoID := r.PathValue("dispositivoID")
	d, err := h.svc.BuscarDispositivo(dispositivoID)
	if err != nil {
		writeErro(w, statusParaErro(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toDispositivoResponse(d))
}

// habilitarDispositivo consulta a biometria e, se aprovada, ativa o dispositivo.
// PATCH /dispositivos/{dispositivoID}/habilitar
// 200 em sucesso; 409 para transição inválida; 422 se biometria reprovada.
func (h *DispositivoHandler) habilitarDispositivo(w http.ResponseWriter, r *http.Request) {
	dispositivoID := r.PathValue("dispositivoID")
	d, err := h.svc.HabilitarDispositivo(dispositivoID)
	if err != nil {
		writeErro(w, statusParaErro(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toDispositivoResponse(d))
}

// cancelarDispositivo faz o soft-delete do dispositivo (transição para cancelado).
// PATCH /dispositivos/{dispositivoID}/cancelar
// 200 em sucesso; idempotente se já estiver cancelado.
func (h *DispositivoHandler) cancelarDispositivo(w http.ResponseWriter, r *http.Request) {
	dispositivoID := r.PathValue("dispositivoID")
	d, err := h.svc.CancelarDispositivo(dispositivoID)
	if err != nil {
		writeErro(w, statusParaErro(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toDispositivoResponse(d))
}

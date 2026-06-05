package service

import (
	"fmt"

	"github.com/balmiza/poc_pm_tech/internal/biometry"
	"github.com/balmiza/poc_pm_tech/internal/domain"
	"github.com/balmiza/poc_pm_tech/internal/repository"
)

// DispositivoService orquestra as regras de negócio do ciclo de vida de dispositivos.
type DispositivoService struct {
	repo      repository.DispositivoRepository
	biometria biometry.Validador
}

func NewDispositivoService(repo repository.DispositivoRepository, b biometry.Validador) *DispositivoService {
	return &DispositivoService{repo: repo, biometria: b}
}

// AutorizacaoResult carrega o resultado da consulta de autorização.
type AutorizacaoResult struct {
	Autorizado    bool
	DispositivoID string
}

// ConsultarAutorizacao verifica se o cliente possui um dispositivo ativo (autorizado).
func (s *DispositivoService) ConsultarAutorizacao(clienteID string) (AutorizacaoResult, error) {
	d, err := s.repo.BuscarAtivoPorClienteID(clienteID)
	if err != nil {
		return AutorizacaoResult{}, fmt.Errorf("consultar autorização: %w", err)
	}
	if d == nil {
		return AutorizacaoResult{Autorizado: false}, nil
	}
	return AutorizacaoResult{Autorizado: true, DispositivoID: d.DispositivoID}, nil
}

// AssociarDispositivo vincula um novo dispositivo ao cliente em estado pendente_confirmacao.
// Pré-condição: o cliente não pode ter outro dispositivo ativo ou pendente.
func (s *DispositivoService) AssociarDispositivo(clienteID string) (*domain.Dispositivo, error) {
	existente, err := s.repo.BuscarAtivoOuPendentePorClienteID(clienteID)
	if err != nil {
		return nil, fmt.Errorf("verificar dispositivo existente: %w", err)
	}
	if existente != nil {
		return nil, domain.ErrClienteJaPossuiDispositivoAtivoPendente
	}

	d, err := domain.NovoDispositivo(clienteID)
	if err != nil {
		return nil, fmt.Errorf("criar dispositivo: %w", err)
	}
	if err := s.repo.Salvar(d); err != nil {
		return nil, fmt.Errorf("salvar dispositivo: %w", err)
	}
	return d, nil
}

// BuscarDispositivo retorna o estado atual de um dispositivo pelo ID.
func (s *DispositivoService) BuscarDispositivo(dispositivoID string) (*domain.Dispositivo, error) {
	d, err := s.repo.BuscarPorID(dispositivoID)
	if err != nil {
		return nil, fmt.Errorf("buscar dispositivo: %w", err)
	}
	return d, nil
}

// HabilitarDispositivo confirma a identidade do cliente via biometria e ativa o dispositivo.
// Pré-condição: dispositivo deve estar em pendente_confirmacao.
// Se a biometria for reprovada, o dispositivo permanece em pendente_confirmacao.
func (s *DispositivoService) HabilitarDispositivo(dispositivoID string) (*domain.Dispositivo, error) {
	d, err := s.repo.BuscarPorID(dispositivoID)
	if err != nil {
		return nil, fmt.Errorf("buscar dispositivo: %w", err)
	}

	if !d.PodeHabilitar() {
		return nil, domain.ErrTransicaoDeEstadoInvalida
	}

	aprovado, err := s.biometria.ValidarBiometria(d.ClienteID)
	if err != nil {
		return nil, fmt.Errorf("validar biometria: %w", err)
	}
	if !aprovado {
		// Dispositivo permanece em pendente_confirmacao conforme regras de negócio.
		return nil, domain.ErrBiometriaReprovada
	}

	if err := d.Habilitar(); err != nil {
		return nil, err
	}
	if err := s.repo.Salvar(d); err != nil {
		return nil, fmt.Errorf("salvar dispositivo: %w", err)
	}
	return d, nil
}

// CancelarDispositivo faz o soft-delete do dispositivo (transição para cancelado).
// É idempotente: se o dispositivo já estiver cancelado, retorna o estado atual sem erro.
func (s *DispositivoService) CancelarDispositivo(dispositivoID string) (*domain.Dispositivo, error) {
	d, err := s.repo.BuscarPorID(dispositivoID)
	if err != nil {
		return nil, fmt.Errorf("buscar dispositivo: %w", err)
	}

	if err := d.Cancelar(); err != nil {
		return nil, err
	}

	if err := s.repo.Salvar(d); err != nil {
		return nil, fmt.Errorf("salvar dispositivo: %w", err)
	}
	return d, nil
}

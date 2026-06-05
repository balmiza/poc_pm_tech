package service_test

import (
	"errors"
	"testing"

	"github.com/balmiza/poc_pm_tech/internal/biometry/mock"
	"github.com/balmiza/poc_pm_tech/internal/domain"
	"github.com/balmiza/poc_pm_tech/internal/repository/memory"
	"github.com/balmiza/poc_pm_tech/internal/service"
)

func novoServico(biometriaAprovada bool) *service.DispositivoService {
	repo := memory.NewDispositivoRepository()
	var bio *mock.Validador
	if biometriaAprovada {
		bio = mock.NovoAprovado()
	} else {
		bio = mock.NovoReprovado()
	}
	return service.NewDispositivoService(repo, bio)
}

// --- AssociarDispositivo ---

func TestAssociarDispositivo_Sucesso(t *testing.T) {
	svc := novoServico(true)
	d, err := svc.AssociarDispositivo("cliente-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Estado != domain.EstadoPendenteConfirmacao {
		t.Errorf("expected %s, got %s", domain.EstadoPendenteConfirmacao, d.Estado)
	}
	if d.DispositivoID == "" {
		t.Error("DispositivoID must not be empty")
	}
}

func TestAssociarDispositivo_ClienteJaPossuiDispositivoPendente(t *testing.T) {
	svc := novoServico(true)
	svc.AssociarDispositivo("cliente-1")

	_, err := svc.AssociarDispositivo("cliente-1")
	if !errors.Is(err, domain.ErrClienteJaPossuiDispositivoAtivoPendente) {
		t.Errorf("expected ErrClienteJaPossuiDispositivoAtivoPendente, got %v", err)
	}
}

func TestAssociarDispositivo_ClienteJaPossuiDispositivoAtivo(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.HabilitarDispositivo(d.DispositivoID)

	_, err := svc.AssociarDispositivo("cliente-1")
	if !errors.Is(err, domain.ErrClienteJaPossuiDispositivoAtivoPendente) {
		t.Errorf("expected ErrClienteJaPossuiDispositivoAtivoPendente, got %v", err)
	}
}

func TestAssociarDispositivo_PermiteAposCancel(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.CancelarDispositivo(d.DispositivoID)

	_, err := svc.AssociarDispositivo("cliente-1")
	if err != nil {
		t.Fatalf("deve permitir nova associação após cancelamento, got: %v", err)
	}
}

func TestAssociarDispositivo_ClientesDiferentesIndependentes(t *testing.T) {
	svc := novoServico(true)
	svc.AssociarDispositivo("cliente-1")

	_, err := svc.AssociarDispositivo("cliente-2")
	if err != nil {
		t.Fatalf("clientes diferentes não devem interferir entre si: %v", err)
	}
}

// --- HabilitarDispositivo ---

func TestHabilitarDispositivo_Sucesso(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	d2, err := svc.HabilitarDispositivo(d.DispositivoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d2.Estado != domain.EstadoAtivo {
		t.Errorf("expected %s, got %s", domain.EstadoAtivo, d2.Estado)
	}
	if d2.DataHabilitacao == nil {
		t.Error("DataHabilitacao must be set")
	}
}

func TestHabilitarDispositivo_BiometriaReprovada_RetornaErro(t *testing.T) {
	svc := novoServico(false)
	d, _ := svc.AssociarDispositivo("cliente-1")

	_, err := svc.HabilitarDispositivo(d.DispositivoID)
	if !errors.Is(err, domain.ErrBiometriaReprovada) {
		t.Errorf("expected ErrBiometriaReprovada, got %v", err)
	}
}

func TestHabilitarDispositivo_BiometriaReprovada_DispositivoPermaneceComPendente(t *testing.T) {
	svc := novoServico(false)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.HabilitarDispositivo(d.DispositivoID)

	// dispositivo ainda está pendente, logo nova associação deve falhar
	_, err := svc.AssociarDispositivo("cliente-1")
	if !errors.Is(err, domain.ErrClienteJaPossuiDispositivoAtivoPendente) {
		t.Errorf("dispositivo deve permanecer pendente após biometria reprovada, got: %v", err)
	}
}

func TestHabilitarDispositivo_NaoEncontrado(t *testing.T) {
	svc := novoServico(true)
	_, err := svc.HabilitarDispositivo("id-inexistente")
	if !errors.Is(err, domain.ErrDispositivoNaoEncontrado) {
		t.Errorf("expected ErrDispositivoNaoEncontrado, got %v", err)
	}
}

func TestHabilitarDispositivo_TransicaoInvalida_Cancelado(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.CancelarDispositivo(d.DispositivoID)

	_, err := svc.HabilitarDispositivo(d.DispositivoID)
	if !errors.Is(err, domain.ErrTransicaoDeEstadoInvalida) {
		t.Errorf("expected ErrTransicaoDeEstadoInvalida, got %v", err)
	}
}

func TestHabilitarDispositivo_TransicaoInvalida_JaAtivo(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.HabilitarDispositivo(d.DispositivoID)

	_, err := svc.HabilitarDispositivo(d.DispositivoID)
	if !errors.Is(err, domain.ErrTransicaoDeEstadoInvalida) {
		t.Errorf("expected ErrTransicaoDeEstadoInvalida, got %v", err)
	}
}

// --- CancelarDispositivo ---

func TestCancelarDispositivo_DePendente(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	d2, err := svc.CancelarDispositivo(d.DispositivoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d2.Estado != domain.EstadoCancelado {
		t.Errorf("expected %s, got %s", domain.EstadoCancelado, d2.Estado)
	}
	if d2.DataCancelamento == nil {
		t.Error("DataCancelamento must be set")
	}
}

func TestCancelarDispositivo_DeAtivo(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.HabilitarDispositivo(d.DispositivoID)
	d2, err := svc.CancelarDispositivo(d.DispositivoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d2.Estado != domain.EstadoCancelado {
		t.Errorf("expected %s, got %s", domain.EstadoCancelado, d2.Estado)
	}
}

func TestCancelarDispositivo_JaCancelado_Idempotente(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.CancelarDispositivo(d.DispositivoID)

	d2, err := svc.CancelarDispositivo(d.DispositivoID)
	if err != nil {
		t.Fatalf("cancelar dispositivo já cancelado deve ser no-op, got: %v", err)
	}
	if d2.Estado != domain.EstadoCancelado {
		t.Errorf("expected %s, got %s", domain.EstadoCancelado, d2.Estado)
	}
}

func TestCancelarDispositivo_NaoEncontrado(t *testing.T) {
	svc := novoServico(true)
	_, err := svc.CancelarDispositivo("id-inexistente")
	if !errors.Is(err, domain.ErrDispositivoNaoEncontrado) {
		t.Errorf("expected ErrDispositivoNaoEncontrado, got %v", err)
	}
}

// --- ConsultarAutorizacao ---

func TestConsultarAutorizacao_DispositivoAtivo_Autorizado(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.HabilitarDispositivo(d.DispositivoID)

	res, err := svc.ConsultarAutorizacao("cliente-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Autorizado {
		t.Error("expected autorizado=true")
	}
	if res.DispositivoID != d.DispositivoID {
		t.Errorf("expected dispositivo_id=%s, got %s", d.DispositivoID, res.DispositivoID)
	}
}

func TestConsultarAutorizacao_SemDispositivo_NaoAutorizado(t *testing.T) {
	svc := novoServico(true)
	res, err := svc.ConsultarAutorizacao("cliente-sem-dispositivo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Autorizado {
		t.Error("expected autorizado=false for client without device")
	}
}

func TestConsultarAutorizacao_DispositivoPendente_NaoAutorizado(t *testing.T) {
	svc := novoServico(true)
	svc.AssociarDispositivo("cliente-1") // apenas associa, sem habilitar

	res, err := svc.ConsultarAutorizacao("cliente-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Autorizado {
		t.Error("dispositivo pendente não deve ser considerado autorizado")
	}
}

func TestConsultarAutorizacao_DispositivoCancelado_NaoAutorizado(t *testing.T) {
	svc := novoServico(true)
	d, _ := svc.AssociarDispositivo("cliente-1")
	svc.HabilitarDispositivo(d.DispositivoID)
	svc.CancelarDispositivo(d.DispositivoID)

	res, err := svc.ConsultarAutorizacao("cliente-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Autorizado {
		t.Error("dispositivo cancelado não deve ser considerado autorizado")
	}
}

package domain

import (
	"testing"
)

func TestNovoDispositivo_EstadoPendenteConfirmacao(t *testing.T) {
	d, err := NovoDispositivo("cliente-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Estado != EstadoPendenteConfirmacao {
		t.Errorf("expected %s, got %s", EstadoPendenteConfirmacao, d.Estado)
	}
	if d.ClienteID != "cliente-1" {
		t.Errorf("expected cliente-1, got %s", d.ClienteID)
	}
	if d.DispositivoID == "" {
		t.Error("DispositivoID should not be empty")
	}
	if d.DataHabilitacao != nil || d.DataCancelamento != nil {
		t.Error("dates should be nil for new device")
	}
}

func TestHabilitar_PendenteParaAtivo(t *testing.T) {
	d, _ := NovoDispositivo("c1")
	if err := d.Habilitar(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Estado != EstadoAtivo {
		t.Errorf("expected %s, got %s", EstadoAtivo, d.Estado)
	}
	if d.DataHabilitacao == nil {
		t.Error("DataHabilitacao should be set after enabling")
	}
}

func TestHabilitar_JaAtivo_Erro(t *testing.T) {
	d, _ := NovoDispositivo("c1")
	d.Habilitar()
	if err := d.Habilitar(); err == nil {
		t.Fatal("expected error when enabling already-active device")
	}
}

func TestHabilitar_Cancelado_Erro(t *testing.T) {
	d, _ := NovoDispositivo("c1")
	d.Cancelar()
	if err := d.Habilitar(); err == nil {
		t.Fatal("expected error when enabling cancelled device")
	}
}

func TestCancelar_PendenteParaCancelado(t *testing.T) {
	d, _ := NovoDispositivo("c1")
	if err := d.Cancelar(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Estado != EstadoCancelado {
		t.Errorf("expected %s, got %s", EstadoCancelado, d.Estado)
	}
	if d.DataCancelamento == nil {
		t.Error("DataCancelamento should be set after cancelling")
	}
}

func TestCancelar_AtivoParaCancelado(t *testing.T) {
	d, _ := NovoDispositivo("c1")
	d.Habilitar()
	if err := d.Cancelar(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Estado != EstadoCancelado {
		t.Errorf("expected %s, got %s", EstadoCancelado, d.Estado)
	}
}

func TestCancelar_JaCancelado_Idempotente(t *testing.T) {
	d, _ := NovoDispositivo("c1")
	d.Cancelar()
	if err := d.Cancelar(); err != nil {
		t.Fatalf("cancelar um dispositivo já cancelado deve ser no-op, got: %v", err)
	}
	if d.Estado != EstadoCancelado {
		t.Errorf("expected %s, got %s", EstadoCancelado, d.Estado)
	}
}

func TestEstaAutorizado_SomenteQuandoAtivo(t *testing.T) {
	d, _ := NovoDispositivo("c1")

	if d.EstaAutorizado() {
		t.Error("pendente device should not be authorized")
	}

	d.Habilitar()
	if !d.EstaAutorizado() {
		t.Error("active device should be authorized")
	}

	d.Cancelar()
	if d.EstaAutorizado() {
		t.Error("cancelled device should not be authorized")
	}
}

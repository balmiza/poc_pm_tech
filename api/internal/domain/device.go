package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"
)

// Estado representa o estado de ciclo de vida de um dispositivo.
type Estado string

const (
	EstadoPendenteConfirmacao Estado = "pendente_confirmacao"
	EstadoAtivo               Estado = "ativo"
	EstadoCancelado           Estado = "cancelado"
)

var (
	ErrDispositivoNaoEncontrado                  = errors.New("dispositivo não encontrado")
	ErrClienteJaPossuiDispositivoAtivoPendente   = errors.New("cliente já possui dispositivo ativo ou pendente de confirmação")
	ErrTransicaoDeEstadoInvalida                 = errors.New("transição de estado inválida para o estado atual")
	ErrBiometriaReprovada                        = errors.New("biometria facial reprovada")
)

// Dispositivo representa o celular de um cliente gerenciado por este serviço.
type Dispositivo struct {
	DispositivoID    string
	ClienteID        string
	Estado           Estado
	DataAssociacao   time.Time
	DataHabilitacao  *time.Time
	DataCancelamento *time.Time
}

// NovoDispositivo cria um dispositivo em estado pendente_confirmacao com ID gerado aleatoriamente.
func NovoDispositivo(clienteID string) (*Dispositivo, error) {
	id, err := gerarID()
	if err != nil {
		return nil, fmt.Errorf("gerar id do dispositivo: %w", err)
	}
	return &Dispositivo{
		DispositivoID:  id,
		ClienteID:      clienteID,
		Estado:         EstadoPendenteConfirmacao,
		DataAssociacao: time.Now().UTC(),
	}, nil
}

func gerarID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// UUID v4 format
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// EstaAutorizado retorna true somente quando o dispositivo está ativo.
func (d *Dispositivo) EstaAutorizado() bool {
	return d.Estado == EstadoAtivo
}

func (d *Dispositivo) PodeHabilitar() bool {
	return d.Estado == EstadoPendenteConfirmacao
}

func (d *Dispositivo) PodeCancelar() bool {
	return d.Estado == EstadoPendenteConfirmacao || d.Estado == EstadoAtivo
}

// Habilitar transiciona o dispositivo de pendente_confirmacao para ativo.
func (d *Dispositivo) Habilitar() error {
	if !d.PodeHabilitar() {
		return ErrTransicaoDeEstadoInvalida
	}
	agora := time.Now().UTC()
	d.Estado = EstadoAtivo
	d.DataHabilitacao = &agora
	return nil
}

// Cancelar transiciona o dispositivo para cancelado. É idempotente: se já estiver
// cancelado, retorna nil sem efeito (regra de negócio: "cancelar um dispositivo já
// cancelado não tem efeito").
func (d *Dispositivo) Cancelar() error {
	if d.Estado == EstadoCancelado {
		return nil
	}
	if !d.PodeCancelar() {
		return ErrTransicaoDeEstadoInvalida
	}
	agora := time.Now().UTC()
	d.Estado = EstadoCancelado
	d.DataCancelamento = &agora
	return nil
}

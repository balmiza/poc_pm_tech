package memory

import (
	"sync"

	"github.com/balmiza/poc_pm_tech/internal/domain"
	"github.com/balmiza/poc_pm_tech/internal/repository"
)

type dispositivoRepository struct {
	mu           sync.RWMutex
	dispositivos map[string]*domain.Dispositivo
}

// NewDispositivoRepository cria um repositório em memória (sem persistência entre
// reinicializações do servidor — adequado para POC sem dependências externas).
func NewDispositivoRepository() repository.DispositivoRepository {
	return &dispositivoRepository{
		dispositivos: make(map[string]*domain.Dispositivo),
	}
}

func (r *dispositivoRepository) Salvar(d *domain.Dispositivo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *d
	r.dispositivos[d.DispositivoID] = &cp
	return nil
}

func (r *dispositivoRepository) BuscarPorID(dispositivoID string) (*domain.Dispositivo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.dispositivos[dispositivoID]
	if !ok {
		return nil, domain.ErrDispositivoNaoEncontrado
	}
	cp := *d
	return &cp, nil
}

func (r *dispositivoRepository) BuscarAtivoOuPendentePorClienteID(clienteID string) (*domain.Dispositivo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.dispositivos {
		if d.ClienteID == clienteID &&
			(d.Estado == domain.EstadoAtivo || d.Estado == domain.EstadoPendenteConfirmacao) {
			cp := *d
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *dispositivoRepository) BuscarAtivoPorClienteID(clienteID string) (*domain.Dispositivo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.dispositivos {
		if d.ClienteID == clienteID && d.Estado == domain.EstadoAtivo {
			cp := *d
			return &cp, nil
		}
	}
	return nil, nil
}

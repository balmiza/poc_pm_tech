package repository

import "github.com/balmiza/poc_pm_tech/internal/domain"

// DispositivoRepository define as operações de persistência de dispositivos.
type DispositivoRepository interface {
	Salvar(d *domain.Dispositivo) error
	BuscarPorID(dispositivoID string) (*domain.Dispositivo, error)
	// BuscarAtivoOuPendentePorClienteID retorna o dispositivo ativo ou pendente do
	// cliente, ou nil se não houver nenhum.
	BuscarAtivoOuPendentePorClienteID(clienteID string) (*domain.Dispositivo, error)
	// BuscarAtivoPorClienteID retorna o dispositivo ativo do cliente, ou nil se não
	// houver nenhum.
	BuscarAtivoPorClienteID(clienteID string) (*domain.Dispositivo, error)
}

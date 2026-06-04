package biometry

// Validador abstrai a API externa de biometria facial.
// O serviço apenas consulta esta interface; não armazena dados biométricos.
type Validador interface {
	// ValidarBiometria retorna (true, nil) se a identidade do cliente foi confirmada,
	// (false, nil) se foi reprovada, ou (false, err) em caso de falha de comunicação.
	ValidarBiometria(clienteID string) (bool, error)
}

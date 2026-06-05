package mock

// Validador é uma implementação fake da API de biometria facial para uso local e testes.
//
// Use NovoAprovado() para simular aprovação (fluxo feliz).
// Use NovoReprovado() para simular rejeição (testar o comportamento quando a biometria falha).
//
// Em main.go a escolha é feita via variável de ambiente BIOMETRIA_RESULTADO.
type Validador struct {
	aprovado bool
}

func NovoAprovado() *Validador  { return &Validador{aprovado: true} }
func NovoReprovado() *Validador { return &Validador{aprovado: false} }

func (v *Validador) ValidarBiometria(_ string) (bool, error) {
	return v.aprovado, nil
}

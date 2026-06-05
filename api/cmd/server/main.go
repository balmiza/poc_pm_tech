package main

import (
	"log"
	"net/http"
	"os"

	"github.com/balmiza/poc_pm_tech/internal/biometry/mock"
	"github.com/balmiza/poc_pm_tech/internal/handler"
	"github.com/balmiza/poc_pm_tech/internal/repository/memory"
	"github.com/balmiza/poc_pm_tech/internal/service"
)

func main() {
	// Biometria: BIOMETRIA_RESULTADO=reprovado rejeita todas as chamadas de biometria.
	// Qualquer outro valor (ou ausência da variável) aprova.
	bio := biometriaFromEnv()

	repo := memory.NewDispositivoRepository()
	svc := service.NewDispositivoService(repo, bio)
	h := handler.NewDispositivoHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := ":" + porta()
	log.Printf("poc_pm_tech iniciado em %s (biometria mock: %s)", addr, os.Getenv("BIOMETRIA_RESULTADO"))
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("erro fatal: %v", err)
	}
}

func biometriaFromEnv() *mock.Validador {
	if os.Getenv("BIOMETRIA_RESULTADO") == "reprovado" {
		return mock.NovoReprovado()
	}
	return mock.NovoAprovado()
}

func porta() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

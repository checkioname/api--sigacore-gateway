package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	// Roteador principal
	mux := chi.NewRouter()

	// Middleware de autenticação que você criaria
	mux.Use(seuMiddlewareDeAuth)

	// Rota para o serviço de clientes (gRPC-Gateway)
	urlClientes, _ := url.Parse("http://localhost:9090")
	proxyClientes := httputil.NewSingleHostReverseProxy(urlClientes)
	mux.Handle("/clientes/*", proxyClientes)

	// Rota para o serviço de relatórios (outra API HTTP)
	urlRelatorios, _ := url.Parse("http://localhost:8081")
	proxyRelatorios := httputil.NewSingleHostReverseProxy(urlRelatorios)
	mux.Handle("/relatorios/*", proxyRelatorios)

	http.ListenAndServe(":8080", mux)
}

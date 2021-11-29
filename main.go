package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/authenticator/oidc"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

func main() {

	dexServerAddress := flag.String("dexServerAddress", "http://127.0.0.1:5556", "dex endpoint")
	url := flag.String("authenticatorUrl", "https://127.0.0.1:8000/", "public endpoint for authenticator")
	dexClientSecret := flag.String("dexClientSecret", "", "dex clinet secret")
	dexCLIClientID := flag.String("dexCLIClientID", "argo-cd", "dex clinet id")
	serveTls := flag.Bool("serveTls", false, "dex clinet id")
	flag.Parse()

	dexConfig := &oidc.DexConfig{
		DexServerAddress: *dexServerAddress,
		Url:              *url,
		DexClientSecret:  *dexClientSecret,
		DexClientID:      *dexCLIClientID,
	}
	oidcClient, dexProxy, err := oidc.GetOidcClient(dexConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	settings, err := oidc.GetSettings(dexConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	sessionManager := middleware.NewSessionManager(settings, dexConfig.DexServerAddress)
	// dex setting ends
	r := mux.NewRouter().StrictSlash(false)
	r.PathPrefix("/api/dex").HandlerFunc(dexProxy)
	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		io.WriteString(writer, "Hello, user!\n")
	})
	r.HandleFunc("/auth/login", oidcClient.HandleLogin)
	r.HandleFunc("/auth/callback", oidcClient.HandleCallback)
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}
	r.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/hello")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8000),
		Handler: middleware.Authorizer(sessionManager)(r),
	}
	if *serveTls {
		cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
		if err != nil {
			log.Fatal(err)
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server.ListenAndServeTLS("", "")
	} else {
		server.ListenAndServe()
	}
	if err != nil {
		log.Fatal(err)
	}
}

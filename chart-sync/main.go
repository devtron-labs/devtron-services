package main

import (
	"github.com/devtron-labs/common-lib/securestore"
	"log"
)

func main() {
	err := securestore.SetEncryptionKey()
	if err != nil {
		log.Println("error in setting encryption key", "err", err)
	}
	app, err := InitializeApp()
	if err != nil {
		log.Panic(err)
	}
	app.Start()
}

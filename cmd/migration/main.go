package main

import (
	"log"

	"github.com/swagftw/covax19-blockchain/pkg/auth"
	"github.com/swagftw/covax19-blockchain/pkg/transaction"
	"github.com/swagftw/covax19-blockchain/pkg/user"
	"github.com/swagftw/covax19-blockchain/utl/storage"
)

func main() {
	// init db.
	gdb, err := storage.NewPostgresDB()
	if err != nil {
		log.Panic(err)
	}

	// create auth schema.
	err = gdb.Exec("CREATE SCHEMA IF NOT EXISTS auth;").Error
	if err != nil {
		log.Panic(err)
	}

	err = gdb.AutoMigrate(&auth.Tokens{})
	if err != nil {
		log.Panic(err)
	}

	// create usr schema.
	err = gdb.Exec("CREATE SCHEMA IF NOT EXISTS usr").Error
	if err != nil {
		log.Panic(err)
	}

	err = gdb.AutoMigrate(&user.User{}, &user.Password{})
	if err != nil {
		log.Panic(err)
	}

	// create transaction schema.
	err = gdb.Exec("CREATE SCHEMA IF NOT EXISTS transactions;").Error
	if err != nil {
		log.Panic(err)
	}

	err = gdb.AutoMigrate(&transaction.Transaction{})
	if err != nil {
		log.Panic(err)
	}
}

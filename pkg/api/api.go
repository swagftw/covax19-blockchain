package api

import (
	"log"

	"github.com/swagftw/covax19-blockchain/pkg/auth"
	authRepo "github.com/swagftw/covax19-blockchain/pkg/auth/repository"
	"github.com/swagftw/covax19-blockchain/pkg/user"
	userRepo "github.com/swagftw/covax19-blockchain/pkg/user/repository"
	authHttp "github.com/swagftw/covax19-blockchain/transport/auth"
	"github.com/swagftw/covax19-blockchain/transport/blockchain"
	userHttp "github.com/swagftw/covax19-blockchain/transport/users"
	"github.com/swagftw/covax19-blockchain/utl/jwt"
	"github.com/swagftw/covax19-blockchain/utl/middleware"
	"github.com/swagftw/covax19-blockchain/utl/server"
	"github.com/swagftw/covax19-blockchain/utl/storage"
	"github.com/swagftw/covax19-blockchain/utl/transaction/postgres"
)

// Start starts the http api proxy server.
func Start() {
	errChan := make(chan error)
	// get echo
	ech := server.InitEcho()
	v1Group := ech.Group("/v1")

	// get db connection
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}

	txn := postgres.NewPostgresTx(db)

	jwtService, err := jwt.New()
	if err != nil {
		log.Fatal(err)
	}

	userService := user.NewService(txn, userRepo.NewRepo(db))
	authService := auth.NewService(txn, authRepo.NewRepo(db), userService, jwtService)

	// init handlers
	userHttp.NewHTTP(v1Group, userService, middleware.JwtMiddleware(jwtService))
	authHttp.NewHTTP(v1Group, authService)
	blockchain.NewHTTP(v1Group, userService)

	// start server
	go func() {
		err := server.StartHTTPServer(ech)
		if err != nil {
			errChan <- err
		}
	}()

	// Wait for the server to exit
	err = <-errChan
	log.Println(err)
}

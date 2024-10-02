package main

import (
	"context"
	"log"
	"time"

	cfg "github.com/dwiw96/simple-auth-system/config"
	factory "github.com/dwiw96/simple-auth-system/factory"
	server "github.com/dwiw96/simple-auth-system/server"
	pg "github.com/dwiw96/simple-auth-system/utils/driver/postgresql"
	rd "github.com/dwiw96/simple-auth-system/utils/driver/redis"

	password "github.com/dwiw96/simple-auth-system/utils/password"
)

func main() {
	log.Println("-- run simple auth system go --")
	env := cfg.GetEnvConfig()
	pgPool := pg.ConnectToPg(env)
	defer pgPool.Close()

	rdClient := rd.ConnectToRedis(env)
	defer rdClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	password.JwtInit(pgPool, ctx)

	router := server.SetupRouter()

	factory.InitFactory(router, pgPool, rdClient, ctx)

	server.StartServer(env.SERVER_PORT, router)
}

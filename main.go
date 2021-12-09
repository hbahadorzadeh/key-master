package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/hbahadorzadeh/key-master/controller/auth"
	"github.com/hbahadorzadeh/key-master/model"
	"github.com/hbahadorzadeh/key-master/service"
	"github.com/hbahadorzadeh/key-master/util"
	log "github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	kill := make(chan os.Signal, 1)
	signal.Notify(kill)

	go func() {
		select {
		case <-kill:
			cancel()
		}
	}()
	app := fx.New(
		fx.Provide(util.NewLogger),
		fx.Provide(util.NewConfigs),
		fx.Provide(service.NewRedisClient),
		fx.Invoke(closeRedis),
		fx.Provide(service.NewValidate),
		fx.Provide(service.NewMongoDatabase),
		fx.Invoke(closeMongodb),
		fx.Invoke(initDatabases),
		fx.Provide(service.NewTokenManager),
		fx.Provide(service.NewWebserver),
		fx.Invoke(initControllers),
		fx.Invoke(runHttpServer),
	)
	if err := app.Start(ctx); err != nil {
		fmt.Println(err)
	}

	//// Match any route
	////app.Use(func(c *fiber.Ctx) error {
	////	fmt.Println("🥇 First handler")
	////	return c.Next()
	////})
	////
	////// Match all routes starting with /api
	////app.Use("/api", func(c *fiber.Ctx) error {
	////	fmt.Println("🥈 Second handler")
	////	return c.Next()
	////})
	////
	////// GET /api/list
	////app.Get("/api/list", func(c *fiber.Ctx) error {
	////	fmt.Println("🥉 Last handler")
	////	return c.SendString("Hello, World 👋!")
	////})

}

func runHttpServer(lifecycle fx.Lifecycle, app *fiber.App, configs *util.Configs) {
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		return app.Listen(fmt.Sprintf("%s:%s", configs.Web.BindAddress, configs.Web.BindPort))
	}})
}

func closeRedis(lifecycle fx.Lifecycle, rdb *redis.Client) {
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		return rdb.Close()
	}})
}

func closeMongodb(lifecycle fx.Lifecycle, mdb *service.MongoDB) {
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		return mdb.Close()
	}})
}

func initControllers(lifecycle fx.Lifecycle, config *util.Configs, logger *log.Logger, app *fiber.App, mdb *service.MongoDB, tokenManager *service.TokenManager) {
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		auth.NewOAuthController(tokenManager).Init(config, logger, app)
		return nil
	}})
}

func initDatabases(lifecycle fx.Lifecycle, mdb *service.MongoDB, logger *log.Logger) {
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		mdb.CreateCollection(model.User{})
		u := &model.User{
			FirstName: "asd",
			LastName:  "dada",
			Email:     "aa@bb.cc",
			IsRemote:  false,
		}
		mdb.Create(u)
		logger.Infof("ID: %s", u.ID)
		u.Password = "asdddddddd"
		mdb.Update(u)
		return nil
	}})
}

package auth

import (
	"fmt"
	jwt "github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/hbahadorzadeh/key-master/service"
	"github.com/hbahadorzadeh/key-master/util"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/shareed2k/goth_fiber"
	log "github.com/sirupsen/logrus"
	"strings"
)

type oAuthProvider struct {
	clientSecret string
	clientId     string
	url          string
	handler      goth.Provider
}

type oAuthController struct {
	tokenManager *service.TokenManager
}

func NewOAuthController(tokenManager *service.TokenManager) (o *oAuthController) {
	return &oAuthController{
		tokenManager: tokenManager,
	}
}
func (o *oAuthController) Init(configs *util.Configs, logger *log.Logger, app *fiber.App) {
	providers := make([]goth.Provider, 0)
	for _, d := range configs.OAuthProviders {
		callback := fmt.Sprintf("%s:%s/auth/callback/%s", configs.Web.ApiBaseUrl, configs.Web.BindPort, strings.ToLower(d.Name))
		switch strings.ToLower(d.Name) {
		case "github":
			providers = append(providers, github.New(d.ClientId, d.ClientSecret, callback))
			logger.Infof("OAuth provider added for `%s`", d.Name)
		case "google":
			providers = append(providers, google.New(d.ClientId, d.ClientSecret, callback))
			logger.Infof("OAuth provider added for `%s`", d.Name)
		}
	}

	goth.UseProviders(providers...)

	app.Get("/auth/login/:provider", goth_fiber.BeginAuthHandler)
	app.Get("/auth/callback/:provider", func(ctx *fiber.Ctx) error {
		user, err := goth_fiber.CompleteUserAuth(ctx)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		provider, err := goth_fiber.GetProviderName(ctx)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		token, err := o.tokenManager.InvokeToken(user, provider)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		return ctx.JSON(fiber.Map{"token": token})
	})

	app.Get("/auth/logout/:provider", func(ctx *fiber.Ctx) error {
		if err := goth_fiber.Logout(ctx); err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		token := ctx.Locals("user").(jwt.Token)
		err := o.tokenManager.RevokeToken(token)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		return ctx.SendString("logout")
	})
}

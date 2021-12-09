package service

import (
	"context"
	"strings"
	"time"

	jwt "github.com/form3tech-oss/jwt-go"
	redis "github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/hbahadorzadeh/key-master/util"
	"github.com/markbates/goth"
	log "github.com/sirupsen/logrus"
)

var ctx = context.Background()

type TokenManager struct {
	logger           *log.Logger
	signingMethod    jwt.SigningMethod
	signingMethodStr string
	signingKey       interface{}
	rdb              *redis.Client
}

func NewTokenManager(configs *util.Configs, rdb *redis.Client, logger *log.Logger) *TokenManager {
	tokenManager := &TokenManager{
		logger:           logger,
		rdb:              rdb,
		signingMethodStr: strings.ToUpper(configs.Web.JwtConfigs.SigningMethod),
	}
	switch tokenManager.signingMethodStr {
	case "ES256":
		tokenManager.signingMethod = jwt.SigningMethodES256
	case "ES384":
		tokenManager.signingMethod = jwt.SigningMethodES384
	case "ES512":
		tokenManager.signingMethod = jwt.SigningMethodES512
	case "HS256":
		tokenManager.signingMethod = jwt.SigningMethodHS256
	case "HS384":
		tokenManager.signingMethod = jwt.SigningMethodHS384
	case "HS512":
		tokenManager.signingMethod = jwt.SigningMethodHS512
	case "RS256":
		tokenManager.signingMethod = jwt.SigningMethodRS256
	case "RS384":
		tokenManager.signingMethod = jwt.SigningMethodRS384
	case "RS512":
		tokenManager.signingMethod = jwt.SigningMethodRS512
	case "PS256":
		tokenManager.signingMethod = jwt.SigningMethodPS256
	case "PS384":
		tokenManager.signingMethod = jwt.SigningMethodPS384
	case "PS512":
		tokenManager.signingMethod = jwt.SigningMethodPS512
	}

	//logger.Infof("Token Key: `%s`", configs.Web.JwtConfigs.SigningKey)
	//if _, err := os.Stat(configs.Web.JwtConfigs.SigningKey); errors.Is(err, os.ErrNotExist) {
	//	pwd, _ := os.Getwd()
	//	logger.Panicf("File not exists!Path: `%s`, PWD: `%s`", configs.Web.JwtConfigs.SigningKey, pwd)
	//}
	//data, err := ioutil.ReadFile(configs.Web.JwtConfigs.SigningKey)
	//if (err!= nil){
	//	logger.Panicf("Error reading pem file: %s", err)
	//}
	private, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(configs.Web.JwtConfigs.SigningKey))
	if err != nil {
		logger.Panicf("Error parsing pem file: %s", err)
	}
	tokenManager.signingKey = private

	return tokenManager
}

func (t *TokenManager) CheckRevokedTokens(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)

	claims := user.Claims.(jwt.MapClaims)
	_, err := t.rdb.Get(ctx, claims["email"].(string)).Result()
	if err == redis.Nil {
		return c.Next()
	} else if err != nil {
		t.logger.Error(err)
	}
	user.Valid = false
	c.Locals("user", user)
	return c.Next()
}

func (t *TokenManager) InvokeToken(user goth.User, provider string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.Name
	claims["email"] = user.Email
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	return token.SignedString(t.signingKey)
}
func (t *TokenManager) InvokeRefreshToken(user goth.User) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 4).Unix()
	return token.SignedString(t.signingKey)
}

func (t *TokenManager) RefreshToken(token jwt.Token, refreshToken jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)
	if time.Unix(claims["exp"].(int64), 0).Before(time.Now()) && time.Unix(refreshClaims["exp"].(int64), 0).Before(time.Now()) {

	}
	//claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 4).Unix()
	return token.SignedString(t.signingKey)
}
func (t *TokenManager) RevokeToken(token jwt.Token) error {
	claims := token.Claims.(jwt.MapClaims)
	timeDiff := time.Unix(claims["exp"].(int64), 0).Sub(time.Now())
	return t.rdb.Set(ctx, claims["email"].(string), 0, timeDiff).Err()
}
func (t *TokenManager) GetMiddleWare() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:     t.signingKey,
		SigningMethod:  t.signingMethodStr,
		SuccessHandler: t.CheckRevokedTokens,
	})
}

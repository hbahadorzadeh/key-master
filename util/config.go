package util

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	ConfigFilePath = "config.json"
)

func NewConfigs(logger *log.Logger) *Configs {
	configs := &Configs{
		DebugMode:      false,
		DB:             &DBConfigs{},
		OAuthProviders: []*OAuthProvider{},
		Web:            &WebConfigs{},
		Mail:           &MailConfigs{},
		Redis:          &RedisConfigs{},
	}
	configs.ParseConfigFile(logger)
	configs.ParseEnvs(logger, os.Environ())
	configs.ParseArgs(logger, os.Args[1:])
	return configs
}

type DBConfigs struct {
	MongoServers []MongoServer `json:"mongo_servers"`
	MongoDB      string        `json:"mongo_db"`
	MongoAuth    MongoAuth     `json:"mongo_auth"`
	PoolSize     int           `json:"pool_size"`
	LogLevel     int           `json:"log_level"`
}

func (configs *Configs) parseDBConfigs(key, value string) {
	switch key {
	case "mongo-servers":
		servers := make([]MongoServer, 0)
		serversStr := strings.Split(value, ",")
		for _, server := range serversStr {
			var port int = 27017
			if strings.Index(server, ":") > 0 {
				portStr, err := strconv.Atoi(strings.Split(server, ":")[1])
				if err == nil {
					port = portStr
				}
			}
			servers = append(servers, MongoServer{
				MongoHost: strings.Split(server, ":")[0],
				MongoPort: port,
			})
		}
		configs.DB.MongoServers = servers
	case "mongo-db":
		configs.DB.MongoDB = value
	case "mongo-auth-mechanism":
		configs.DB.MongoAuth.AuthMechanism = value
	case "mongo-auth-mechanism-properties":
		properties := make(map[string]string)
		for _, v := range strings.Split(value, ",") {
			if strings.Index(v, ":") > 0 {
				properties[strings.Split(v, ":")[0]] = strings.Split(v, ":")[1]
			}
		}
		configs.DB.MongoAuth.AuthMechanismProperties = properties
	case "mongo-auth-source":
		configs.DB.MongoAuth.AuthSource = value
	case "mongo-auth-username":
		configs.DB.MongoAuth.Username = value
	case "mongo-auth-password":
		configs.DB.MongoAuth.Password = value
	case "pool-size":
		poolSize, err := strconv.Atoi(value)
		if err == nil {
			configs.DB.PoolSize = poolSize
		}
	case "log-level":
		logLevel, err := strconv.Atoi(value)
		if err == nil {
			configs.DB.LogLevel = logLevel
		}
	}
}

type MongoServer struct {
	MongoHost string `json:"mongo_host"`
	MongoPort int    `json:"mongo_port"`
}

type MongoAuth struct {
	AuthMechanism           string            `json:"auth_mechanism"`
	AuthMechanismProperties map[string]string `json:"auth_mechanism_properties"`
	AuthSource              string            `json:"auth_source"`
	Username                string            `json:"username"`
	Password                string            `json:"password"`
}

type OAuthProvider struct {
	ClientSecret string `json:"client_secret"`
	ClientId     string `json:"client_token"`
	Name         string `json:"name"`
}

func (configs *Configs) parseOAuthProvider(key, value string) {
	//switch key {
	//case "client-secret":
	//	configs.OAuthProviders.
	//}
}

type JwtConfigs struct {
	SigningKey    string `json:"signing_key"`
	SigningMethod string `json:"signing_method"`
	SigningKeyPW  string `json:"signing_key_pw"`
}

type WebConfigs struct {
	ApiBaseUrl  string `json:"api_base_url"`
	BindAddress string `json:"bind_address"`
	BindPort    string `json:"bind_port"`
	UiUrl       string
	JwtConfigs  JwtConfigs `json:"jwt_configs"`
}

func (configs *Configs) parseWebConfigs(key, value string) {
	switch key {
	case "api-base-url":
		configs.Web.ApiBaseUrl = value
	case "bind-address":
		configs.Web.BindAddress = value
	case "bind-port":
		configs.Web.BindPort = value
	case "jwt-configs-signing-key":
		configs.Web.JwtConfigs.SigningKey = value
	case "jwt-configs-signing-method":
		configs.Web.JwtConfigs.SigningMethod = value
	}
}

type MailConfigs struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Host          string `json:"host"`
	SenderAddress string `json:"sender-address"`
	Port          int    `json:"port"`
}

type RedisConfigs struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	Index    int    `json:"index"`
}

func (configs *Configs) parseRedisConfigs(key, value string) {
	switch key {
	case "address":
		configs.Redis.Address = value
	case "password":
		configs.Redis.Password = value
	case "Index":
		var index int = 0
		indexStr, err := strconv.Atoi(value)
		if err == nil {
			index = indexStr
		}
		configs.Redis.Index = index
	}
}

type Configs struct {
	DebugMode      bool             `json:"debug_mode"`
	DB             *DBConfigs       `json:"db"`
	OAuthProviders []*OAuthProvider `json:"o_auth_providers"`
	Web            *WebConfigs      `json:"web"`
	Mail           *MailConfigs     `json:"mail"`
	Redis          *RedisConfigs    `json:"redis"`
}

func (configs *Configs) ParseConfigFile(logger *log.Logger) {
	if file, err := ioutil.ReadFile(ConfigFilePath); err == nil {
		json.Unmarshal([]byte(file), configs)
	} else {
		logger.Printf("Failed to read config file: `%i`", err)
	}
}

func parseArg(arg string) (section, key, value string) {
	arg = strings.TrimLeft(arg, "--")
	keyValueStr := ""
	if strings.Index(arg, "-") == -1 {
		if strings.Index(arg, "=") >= 0 {
			section = arg[:strings.Index(arg, "=")]
			keyValueStr = arg[strings.Index(arg, "="):]
		} else {
			section = arg
			keyValueStr = "="
		}
	} else {
		section = arg[:strings.Index(arg, "-")]
		keyValueStr = arg[strings.Index(arg, "-")+1:]
	}

	keyValue := strings.Split(keyValueStr, "=")
	key = keyValue[0]
	value = keyValue[1]
	return
}

func (configs *Configs) setConfigs(section, key, value string) {
	switch strings.ToLower(section) {
	case "debug":
		configs.DebugMode = strings.ToLower(value) == "true"
	case "web":
		configs.parseWebConfigs(key, value)
	case "db":
		configs.parseDBConfigs(key, value)
	case "redis":
		configs.parseRedisConfigs(key, value)
	}
}

func (configs *Configs) ParseEnvs(logger *log.Logger, envs []string) {
	var section, key, value string
	for _, env := range envs {
		if strings.Index(env, "_") < 0 {
			continue
		}
		section = strings.ToLower(strings.Split(env, "_")[0])
		keyValue := strings.Split(strings.ReplaceAll(strings.ToLower(env[strings.Index(env, "_")+1:]), "_", "-"), "=")
		if len(keyValue) <= 1 {
			continue
		}
		key = keyValue[0]
		value = keyValue[1]
		logger.Infof("Sec: `%s`, Key `%s`, Value: `%s`", section, key, value)
		configs.setConfigs(section, key, value)
	}
	logger.Printf("`%d` env processed", len(envs))
}

func (configs *Configs) ParseArgs(logger *log.Logger, args []string) {
	for _, arg := range args {
		section, key, value := parseArg(arg)
		logger.Infof("Sec: `%s`, Key `%s`, Value: `%s`", section, key, value)
		// arg are formatted like --db-key=value
		configs.setConfigs(section, key, value)
	}
	logger.Printf("`%d` args processed", len(args))
}

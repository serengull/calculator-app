package configuration

import (
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigurationManager interface {
	GetServerConfig() ServerConfig
}

type configurationManagerImp struct {
	mu                sync.RWMutex
	applicationConfig ApplicationConfig
}

func (configurationManager *configurationManagerImp) GetServerConfig() ServerConfig {
	configurationManager.mu.RLock()
	defer configurationManager.mu.RUnlock()

	return configurationManager.applicationConfig.Server
}

// setApplicationConfig swaps the served config, so a reload actually takes hold.
func (configurationManager *configurationManagerImp) setApplicationConfig(conf ApplicationConfig) {
	configurationManager.mu.Lock()
	defer configurationManager.mu.Unlock()

	configurationManager.applicationConfig = conf
}

var instance *configurationManagerImp

const (
	serverPort = ":8080"
	configType = "yaml"
)

func NewConfigurationManager(configPath string, configName string) ConfigurationManager {
	if instance != nil {
		return instance
	}

	env := os.Getenv("ACTIVE_PROFILE")
	if env == "" {
		log.Print("**** ACTIVE_PROFILE is empty, default it will be used as 'dev' ****")
		env = "dev"
	}

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)
	viper.SetTypeByDefaultValue(true)
	viper.SetDefault("server.port", serverPort)
	viper.WatchConfig()

	instance = &configurationManagerImp{applicationConfig: readConf(env, configPath, configName)}

	viper.OnConfigChange(func(e fsnotify.Event) {
		instance.setApplicationConfig(readConf(env, configPath, configName))
		logrus.WithField("file", e.Name).Warn("Config file changed")
	})

	return instance
}

func readConf(env string, configPath string, configName string) ApplicationConfig {
	readConfigErr := viper.ReadInConfig()
	if readConfigErr != nil {
		log.Panicf("Couldn't load application configuration, cannot start. Error details: %s", readConfigErr.Error())
	}
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)
	viper.SetTypeByDefaultValue(true)
	mergeConfigErr := viper.MergeInConfig()
	if mergeConfigErr != nil {
		log.Panicf("Couldn't load application configuration, cannot start. Error details: %s", mergeConfigErr.Error())
	}
	var conf ApplicationConfig
	c := viper.Sub(env)
	if c == nil {
		log.Panicf("No configuration found for ACTIVE_PROFILE %q in %s/%s.%s. Terminating.", env, configPath, configName, configType)
	}
	if unMarshalErr := c.Unmarshal(&conf); unMarshalErr != nil {
		log.Panicf("Configuration cannot deserialize. Terminating. Error details: %s", unMarshalErr.Error())
	}

	logrus.WithField("configuration", conf).Debug("Configuration changed")

	return conf
}

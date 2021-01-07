package config

import (
	"clx/constants/settings"
	"clx/file"
	"clx/structs"
	"github.com/spf13/viper"
)

func GetConfig() *structs.Config {
	viper.SetConfigName(settings.ConfigFileNameAbbreviated)
	viper.AddConfigPath(file.PathToConfigDirectory())
	viper.AutomaticEnv()
	viper.SetConfigType("env")

	configuration := new(structs.Config)
	_ = viper.ReadInConfig()

	setDefaultValues()

	err := viper.Unmarshal(&configuration)
	if err != nil {
		panic(err)
	}

	return configuration
}

func setDefaultValues() {
	viper.SetDefault(settings.CommentWidthKey, settings.CommentWidthDefault)
	viper.SetDefault(settings.IndentSizeKey, settings.IndentSizeDefault)
}

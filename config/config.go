package config

import (
	"clx/constants/settings"
	"clx/file"
	"clx/structs"
	"fmt"
	"github.com/spf13/viper"
)

func GetConfig() *structs.Config {
	// Set the file name of the configurations file
	viper.SetConfigName(settings.ConfigFileNameAbbreviated)

	cp := file.PathToConfigDirectory()
	viper.AddConfigPath(cp)

	//Check for environment variables
	viper.AutomaticEnv()

	viper.SetConfigType("env")

	configuration := new(structs.Config)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	setDefaultValues()

	err := viper.Unmarshal(&configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	return configuration
}

func setDefaultValues() {
	viper.SetDefault(settings.CommentWidthKey, settings.CommentWidthDefault)
	viper.SetDefault(settings.IndentSizeKey, settings.IndentSizeDefault)
}

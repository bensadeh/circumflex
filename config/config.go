package config

import (
	"clx/core"
	"clx/file"
	"clx/settings"

	"github.com/spf13/viper"
)

func GetConfig() *core.Config {
	viper.SetConfigName(settings.ConfigFileNameAbbreviated)
	viper.AddConfigPath(file.PathToConfigDirectory())
	viper.AutomaticEnv()
	viper.SetConfigType("env")

	configuration := new(core.Config)
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
	viper.SetDefault(settings.PreserveRightMarginKey, settings.PreserveRightMarginDefault)
	viper.SetDefault(settings.HighlightHeadlinesKey, settings.HighlightHeadlinesDefault)
	viper.SetDefault(settings.RelativeNumberingKey, settings.RelativeNumberingDefault)
	viper.SetDefault(settings.HideYCJobsKey, settings.HideYCJobsDefault)
	viper.SetDefault(settings.UseAltIndentBlockKey, settings.UseAltIndentBlockDefault)
	viper.SetDefault(settings.HighlightCommentsKey, settings.HighlightCommentsDefault)
	viper.SetDefault(settings.EmojiSmileysKey, settings.EmojiSmileysDefault)
	viper.SetDefault(settings.MarkAsReadKey, settings.MarkAsReadDefault)
	viper.SetDefault(settings.HideIndentSymbolKey, settings.HideIndentSymbolDefault)
	viper.SetDefault(settings.OrangeHeaderKey, settings.OrangeHeaderDefault)
}

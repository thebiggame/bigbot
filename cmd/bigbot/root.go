package main

import (
	"github.com/thebiggame/bigbot/cmd/bigbot/run"
	"github.com/thebiggame/bigbot/internal/config"
	"github.com/thebiggame/bigbot/internal/log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "bigbot",
		Short: "The theBIGGAME Discord / Automations bot.",
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./bigbot.yaml)")

	rootCmd.PersistentFlags().String("log.level", "info", "Log level (trace, debug, info, warn, error, fatal)")

	viper.BindPFlag("log.level", rootCmd.Flags().Lookup("log.level"))

	rootCmd.AddCommand(run.Cmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		pwd, err := os.Getwd()
		cobra.CheckErr(err)

		viper.AddConfigPath(pwd)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".bigbot")
	}

	viper.SetEnvPrefix("BIGBOT_")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}

	config.BindViperConfig()

	if err := viper.Unmarshal(&config.RuntimeConfig); err != nil {
		cobra.CheckErr(err)
	}
}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port string `mapstructure:"port"`
}
type LineConfig struct {
	Secret string `mapstructure:"secret"`
	Token  string `mapstructure:"token"`
}
type DBConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
}
type Config struct {
	ServerConfig `mapstructure:"server"`
	LineConfig   `mapstructure:"line"`
	DBConfig     `mapstructure:"db"`
}

var (
	configName string
	config     Config
)

var rootCmd = &cobra.Command{
	Use:   "messenger",
	Short: "messenger - a http server that integrates Line and MongoDB",
	Long: `The messenger is a http server that integrates Line and MongoDB.
    You can use config, environment variables or CLI flags to set basic configuration to connect to Line and MongoDB.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVarP(&configName, "config", "c", "", "config file(it should be placed in the root path)")
	rootCmd.Flags().StringP("sever_port", "", "8085", "server port")
	rootCmd.Flags().StringP("line_secret", "", "", "channel secret of Line")
	rootCmd.Flags().StringP("line_token", "", "", "channel access token of Line")

	rootCmd.Flags().StringP("db_user", "u", "", "database user")
	rootCmd.Flags().StringP("db_password", "p", "", "database password")
	rootCmd.Flags().StringP("db_host", "", "localhost", "database host")
	rootCmd.Flags().StringP("db_port", "", "27017", "database port")

	viper.BindPFlag("sever.port", rootCmd.Flags().Lookup("sever_port"))
	viper.BindPFlag("line.token", rootCmd.Flags().Lookup("line_secret"))
	viper.BindPFlag("line.token", rootCmd.Flags().Lookup("line_token"))

	viper.BindPFlag("db.user", rootCmd.Flags().Lookup("db_user"))
	viper.BindPFlag("db.password", rootCmd.Flags().Lookup("db_password"))
	viper.BindPFlag("db.host", rootCmd.Flags().Lookup("db_host"))
	viper.BindPFlag("db.port", rootCmd.Flags().Lookup("db_port"))
}

func initConfig() {
	if configName != "" {
		viper.SetConfigName(configName)
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("..")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Can not read config: = %v\n", err)
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
}

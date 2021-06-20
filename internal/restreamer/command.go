package restreamer

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use: "restreamer",
}

func Main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("restreamer")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".restreamer"))
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("restreamer.yaml not found...exiting")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	rootCmd.PersistentFlags().Float64P("max-bandwidth", "m", 10, "max bandwidth in mb/sec")
	rootCmd.PersistentFlags().Float64P("read-buffer", "b", 1, "read buffer in MB")

	bindFlagToConfig(rootCmd, "max-bandwidth", "max-bandwidth")
	bindFlagToConfig(rootCmd, "read-buffer", "read-buffer")
}

func bindFlagToConfig(cmd *cobra.Command, flag, configPath string) {
	pflag := cmd.Flags().Lookup(flag)
	if pflag == nil {
		pflag = cmd.PersistentFlags().Lookup(flag)
	}
	if pflag == nil {
		log.Fatalf("flag %s not found", flag)
	}

	if err := viper.BindPFlag(configPath, pflag); err != nil {
		log.Fatal(err)
	}
}

/*
Copyright © 2021 Urjit Singh Bhatia

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "buckit",
	Short: "A simple s3 front-end server",
	Long: `Keep your buckets secure by routing requests through buckit:
	No need to expose your buckets to the public or in website mode`,
	Run: func(cmd *cobra.Command, args []string) {
		fetcher := App{}
		err := viper.Unmarshal(&fetcher.Config)
		if err != nil {
			log.Fatalln("couldn't parse config file")
		}
		log.Println("Starting...")

		viper.OnConfigChange(func(in fsnotify.Event) {
			log.Printf("Dynamically updating config. File changed: %s", in.Name)
			newCfg := &config{}
			err := viper.Unmarshal(newCfg)
			if err != nil {
				log.Printf("couldn't parse updated config file - changed won't take effect. Error: %s", err)
				return
			}
			log.Println("Updated config")
			// only change the dynamic bits
			fetcher.Config.Buckits = newCfg.Buckits
			fetcher.Config.ShutdownTimeout = newCfg.ShutdownTimeout
		})
		viper.WatchConfig()

		sigChan := make(chan os.Signal)
		go registerSignals(sigChan, &fetcher)

		fetcher.Start()

		<-sigChan
		log.Println("Shutdown complete")
	},
}

func registerSignals(sigChan chan os.Signal, fetcher *App) {
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Captured signal: %s. Stopping gracefully. Timeout: %s", sig, fetcher.Config.ShutdownTimeout)
	if err := fetcher.Stop(); err != nil {
		log.Fatalf("Did not shutdown gracefully. %s", err)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .buckit.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".buckit" (without extension).
		viper.AddConfigPath("./")
		viper.SetConfigName(".buckit")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Fatalln("Cannot start without a config file")
	}
}

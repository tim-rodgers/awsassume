// Copyright © 2019 Timothy Rodgers <rodgers.timothy2gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFileFlag string
var profileNameFlag string
var configPathFlag string
var credsPathFlag string
var durationFlag int
var loggingLevelFlag string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awsassume",
	Short: "A tool to make assuming AWS roles easier",
	Long: `awsassume allows you to run commands or start a new shell with temporary
credentials sourced from the AWS STS API.

See https://github.com/tim-rodgers/awsassume for documentation`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.Execute()
}

func init() {
	bindEnvironment()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&profileNameFlag, "profile", "p", "", "profile to be assumed (required)")
	viper.BindPFlag("Profile", rootCmd.PersistentFlags().Lookup("profile"))
	rootCmd.PersistentFlags().StringVar(&configPathFlag, "aws-config-file", "~/.aws/config", "Path to AWS CLI config file")
	rootCmd.PersistentFlags().StringVar(&credsPathFlag, "aws-credentials-file", "~/.aws/credentials", "Path to AWS shared credentials file")
	rootCmd.PersistentFlags().IntVarP(&durationFlag, "duration", "d", 15, "How long in minutes credentials should be valid for")
	rootCmd.PersistentFlags().StringVarP(&loggingLevelFlag, "log-level", "l", "info", "logging level")
	viper.BindPFlag("SessionDuration", rootCmd.PersistentFlags().Lookup("duration"))
	viper.BindPFlag("AWSSharedCredentialsFile", rootCmd.PersistentFlags().Lookup("aws-credentials-file"))
	viper.BindPFlag("AWSConfigFile", rootCmd.PersistentFlags().Lookup("aws-config-file"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFileFlag != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFileFlag)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".awsassume" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".awsassume")
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }
	level, _ := log.ParseLevel(loggingLevelFlag)
	log.SetLevel(level)
}

func bindEnvironment() {
	viper.BindEnv("SessionDuration", "AWSASSUME_DURATION")
	viper.BindEnv("AWSSharedCredentialsFile", "AWS_SHARED_CREDENTIALS_FILE")
	viper.BindEnv("AWSConfigFile", "AWS_CONFIG_FILE")
	viper.BindEnv("Region", "AWS_DEFAULT_REGION")
	viper.BindEnv("SourceProfile", "AWS_PROFILE")
}

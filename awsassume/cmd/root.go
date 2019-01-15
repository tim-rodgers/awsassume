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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFileFlag string
var commandFlag string
var configPathFlag string
var credsPathFlag string
var profileNameFlag string
var durationFlag int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "awsassume",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
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
	bindEnvironment()
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFileFlag, "config", "", "Config file (default is $HOME/.awsassume.yaml)")
	rootCmd.PersistentFlags().StringVar(&configPathFlag, "aws-config-file", "~/.aws/config", "Path to AWS CLI config file")
	rootCmd.PersistentFlags().StringVar(&credsPathFlag, "aws-credentials-file", "~/.aws/credentials", "Path to AWS shared credentials file")
	rootCmd.PersistentFlags().StringVarP(&commandFlag, "command", "c", os.Getenv("SHELL"), "Command to use")
	rootCmd.PersistentFlags().IntVarP(&durationFlag, "duration", "d", 15, "How long credentials should be valid for")
	rootCmd.PersistentFlags().StringVarP(&profileNameFlag, "profile", "p", "default", "Profile to assume (Required)")
	rootCmd.PersistentFlags().StringVarP(&profileNameFlag, "source-profile", "s", "default", "Source profile for credentials")
	rootCmd.MarkPersistentFlagRequired("profile")
	viper.BindPFlag("DefaultCommand", rootCmd.PersistentFlags().Lookup("command"))
	viper.BindPFlag("DefaultDuration", rootCmd.PersistentFlags().Lookup("duration"))
	viper.BindPFlag("DefaultSourceProfile", rootCmd.PersistentFlags().Lookup("source-profile"))
	viper.BindPFlag("AWSSharedCredentialsFile", rootCmd.PersistentFlags().Lookup("aws-credentials-file"))
	viper.BindPFlag("AWSConfigFile", rootCmd.PersistentFlags().Lookup("aws-config-file"))
	viper.BindPFlag("AWSDefaultRegion", rootCmd.PersistentFlags().Lookup("region"))
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
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func bindEnvironment() {
	viper.BindEnv("DefaultCommand", "AWSASSUME_COMMAND")
	viper.BindEnv("DefaultDuration", "AWSASSUME_DURATION")
	viper.BindEnv("AWSSharedCredentialsFile", "AWS_SHARED_CREDENTIALS_FILE")
	viper.BindEnv("AWSConfigFile", "AWS_CONFIG_FILE")
	viper.BindEnv("AWSDefaultRegion", "AWS_DEFAULT_REGION")
	viper.BindEnv("AWSDefaultSourceProfile", "AWS_PROFILE")
}

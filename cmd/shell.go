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
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell",
	Args:  cobra.NoArgs,
	Short: "Start a shell session with an assumed role",
	Long:  `Fetches temporary credentials and starts a new shell with the credentials set as env vars.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := shell(); err != nil {
			log.WithError(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func shell() error {
	if viper.GetString("Profile") == "" {
		rootCmd.Help()
		return errors.New("profile must be provided")
	}
	if os.Getenv("AWSASSUME") != "" {
		return errors.New("In an awsassume shell. Exit this before running further commands")
	}
	command := []string{os.Getenv("SHELL")}
	return ExecuteCommand(command)
}

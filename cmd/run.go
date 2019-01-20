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

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Args:  cobra.MinimumNArgs(1),
	Short: "Run a single command with assumed role",
	Long:  `Assumes a role and uses the returned credentials to execute a single command`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(args); err != nil {
			log.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
func run(args []string) error {
	if viper.GetString("Profile") == "" {
		rootCmd.Help()
		return errors.New("profile must be provided")
	}
	if os.Getenv("AWSASSUME") != "" {
		return errors.New("In an awsassume shell. Exit this before running further commands")
	}
	return ExecuteCommand(args)
}

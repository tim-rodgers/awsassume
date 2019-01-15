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
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var shellCommand string

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start a shell session with an assumed role",
	Long:  `Fetches temporary credentials and starts a new shell with the credentials set as env vars.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := shell(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
	rootCmd.Flags().StringVarP(&shellCommand, "shell-command", "s", os.Getenv("SHELL"), "shell command (default $SHELL)")
}

func shell() error {
	if os.Getenv("AWSASSUME") != "" {
		return errors.New("In an awsassume shell. Exit this before running further commands")
	}
	cmd := exec.Command(shellCommand)
	cmd.Env = EnvVars()
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	return nil
}

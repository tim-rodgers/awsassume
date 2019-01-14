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
	"strings"

	"github.com/tim-rodgers/awsassume"

	"github.com/spf13/cobra"
)

// InputCommand contains data related to the command to run
type InputCommand struct {
	Profile string
	Command string
	Args    []string
}

var profile string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use: "run",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must provide a command to execute")
		}
		return nil
	},
	Short: "run a single command with assumed role",
	Long: `Assumes a role and uses the returned credentials
to execute a single command`,
	Run: func(cmd *cobra.Command, args []string) {
		inputCmd := InputCommand{
			Profile: profile,
			Command: command,
			Args:    args,
		}
		if err := run(inputCmd); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&profile, "profile", "p", "", "Profile to assume")
	runCmd.MarkFlagRequired("profile")
}

func run(inputCmd InputCommand) error {
	if os.Getenv("AWSASSUME") != "" {
		return errors.New("In an awsassume shell. Exit this before running further commands")
	}
	credentials, err := awsassume.Get(awsConfigPath, awsCredsPath, profile, sourceProfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// prepend `-c` switch to args passed to exec.Command
	cmdArgs := []string{"-c", strings.Join(inputCmd.Args, " ")}
	fmt.Println("args are: ", cmdArgs)
	cmd := exec.Command(inputCmd.Command, cmdArgs...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AWSASSUME=1"),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credentials.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credentials.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credentials.SessionToken),
		fmt.Sprintf("AWS_DEFAULT_REGION=%s", credentials.Region),
		fmt.Sprintf("AWSASSUME_EXPIRY=%s", credentials.ExpiresAt),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return nil
}

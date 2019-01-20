package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tim-rodgers/awsassume/awsassume"
)

// EnvVars returns a slice containing the current environment variables and the variables
// to set based on the assumed profile
func EnvVars(credentials *awsassume.CredentialsValue, region string) (envVars []string) {
	envVars = append(envVars,
		fmt.Sprintf("AWSASSUME=1"),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credentials.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credentials.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credentials.SessionToken),
		fmt.Sprintf("SESSION_EXPIRATION=%s", credentials.SessionExpiration),
	)
	if cRegion := viper.Get("Region"); cRegion != nil {
		envVars = append([]string{fmt.Sprintf("AWS_DEFAULT_REGION=%s", cRegion)}, envVars...)
	} else if region != "" {
		envVars = append([]string{fmt.Sprintf("AWS_DEFAULT_REGION=%s", region)}, envVars...)
	}
	envVars = append(os.Environ(), envVars...)
	return envVars
}

// ExecuteCommand executes a command with the credentials specified
func ExecuteCommand(args []string) error {
	command := args[0]
	credentialsClient, err := awsassume.NewCredentialsClient(
		viper.GetString("AWSConfigFile"),
		viper.GetString("AWSSharedCredentialsFile"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	profile, err := credentialsClient.ConfigProvider.GetProfile(viper.GetString("Profile"))
	if err != nil {
		log.WithError(err)
		os.Exit(1)
	}
	options := awsassume.AssumeRoleOptions{
		ProfileName:     viper.GetString("Profile"),
		SourceProfile:   profile.SourceProfile,
		RoleARN:         profile.RoleArn,
		MFASerial:       profile.MfaSerial,
		ExternalID:      profile.ExternalID,
		RoleSessionName: profile.RoleSessionName,
		SessionDuration: time.Duration(viper.GetInt("SessionDuration")),
	}
	credentials, err := credentialsClient.GetCredentials(options)
	if err != nil {
		log.WithError(err)
		os.Exit(1)
	}
	binary, err := exec.LookPath(command)
	if err != nil {
		log.WithError(err)
		os.Exit(1)
	}
	return syscall.Exec(binary, args, EnvVars(credentials, profile.Region))
}

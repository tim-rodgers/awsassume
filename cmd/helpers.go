package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/tim-rodgers/awsassume/awsassume"
)

// FetchCredentials sets the credentials for the command line tool
func FetchCredentials() *awsassume.Value {
	credentialProvider := awsassume.CredentialProvider{
		ConfigFile:    viper.GetString("AWSConfigFile"),
		CredsFile:     viper.GetString("AWSSharedCredentialsFile"),
		ProfileName:   viper.GetString("ProfileName"),
		SourceProfile: viper.GetString("SourceProfile"),
		Duration:      viper.GetInt("SessionDuration"),
		Region:        viper.GetString("Region"),
	}
	var err error
	credentials, err := credentialProvider.Retrieve()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return credentials
}

// EnvVars returns a slice containing the current environment variables and the variables
// to set based on the assumed profile
func EnvVars(val *awsassume.Value) []string {
	envVars := []string{
		fmt.Sprintf("AWSASSUME=1"),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", val.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", val.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", val.SessionToken),
		fmt.Sprintf("AWSASSUME_EXPIRY=%s", val.SessionExpiration),
	}
	if val.Region != "" {
		envVars = append([]string{fmt.Sprintf("AWS_DEFAULT_REGION=%s", val.Region)}, envVars...)
	}
	envVars = append(os.Environ(), envVars...)
	return envVars
}

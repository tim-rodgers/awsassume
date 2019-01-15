package cmd

import (
	"fmt"
	"os"
)

// EnvVars returns a slice containing the current environment variables and the variables
// to set based on the assumed profile
func EnvVars() []string {
	envVars := []string{
		fmt.Sprintf("AWSASSUME=1"),
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credentials.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credentials.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credentials.SessionToken),
		fmt.Sprintf("AWSASSUME_EXPIRY=%s", credentials.SessionExpiration),
	}
	if credentials.Region != "" {
		envVars = append([]string{fmt.Sprintf("AWS_DEFAULT_REGION=%s", credentials.Region)}, envVars...)
	}
	envVars = append(os.Environ(), envVars...)
	return envVars
}

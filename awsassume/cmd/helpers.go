package cmd

import "os"

// EnvConfig holds the value of supported environment variables
type EnvConfig struct {
	SharedCredentialsFile string
	ConfigFile            string
	Region                string
	SourceProfile         string
	AwsAssume             string
}

// GetEnvConfig retrieves the values of configuration values that can be set by environment
// variables
func GetEnvConfig() *EnvConfig {
	envConfig := new(EnvConfig)
	envConfig.SharedCredentialsFile = os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	envConfig.ConfigFile = os.Getenv("AWS_CONFIG_FILE")
	envConfig.Region = os.Getenv("AWS_DEFAULT_REGION")
	envConfig.SourceProfile = os.Getenv("AWS_DEFAULT_PROFILE")
	envConfig.AwsAssume = os.Getenv("AWSASSUME")
	return envConfig
}

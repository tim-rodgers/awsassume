// Copyright © 2019 Timothy Rodgers <rodgers.timothy@gmail.com>
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

package awsassume

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// ConfigProvider is an interface to retrieve config and credentials stored
// locally
type ConfigProvider interface {
	GetProfile(profileName string) (*ProfileConfig, error)
	GetCredentials(profileName string) (*CredentialsValue, error)
	SetCredentials(profileName string, credentials *CredentialsValue) error
}

// CredentialsProvider is an interface to retrieve temporary credentials for a profile
// in the AWS config file
type CredentialsProvider interface {
	AssumeRole(options AssumeRoleOptions) (*CredentialsValue, error)
}

// ProfileConfig contains the properties for a profile stored in the config file
type ProfileConfig struct {
	SourceProfile   string `ini:"source_profile"`
	RoleArn         string `ini:"role_arn"`
	MfaSerial       string `ini:"mfa_serial"`
	ExternalID      string `ini:"external_id"`
	Region          string `ini:"region"`
	RoleSessionName string `ini:"role_session_name"`
}

// CredentialsValue represents the temporary credentials returned by AWS or read
// from the credentials file
type CredentialsValue struct {
	AccessKeyID       string    `ini:"aws_access_key_id"`
	SecretAccessKey   string    `ini:"aws_secret_access_key"`
	SessionToken      string    `ini:"aws_session_token"`
	SessionExpiration time.Time `ini:"aws_session_expiration"`
}

// AssumeRoleOptions holds the configurations values to be passed to the AssumeRole
// function
type AssumeRoleOptions struct {
	ProfileName     string
	SourceProfile   string
	RoleARN         string
	MFASerial       string
	ExternalID      string
	RoleSessionName string
	SessionDuration time.Duration
}

// AWSConfigProvider Fetches profile and credential data from aws configuration files
type AWSConfigProvider struct {
	configPath         string
	credentialsPath    string
	awsConfigFile      *ini.File
	awsCredentialsFile *ini.File
}

// GetProfile retrieves a profile from the config file if it exists, or an error
// if no profile is found
func (c *AWSConfigProvider) GetProfile(profileName string) (*ProfileConfig, error) {
	log.Debugf("Looking up profile %s in config file", profileName)
	var sectionName string
	if profileName == "default" {
		sectionName = profileName
	} else {
		sectionName = fmt.Sprintf("profile %s", profileName)
	}
	if section := c.awsConfigFile.Section(sectionName); section != nil {
		log.Debugf("Profile %s found", profileName)
		var profile = &ProfileConfig{}
		section.MapTo(profile)
		return profile, nil
	}
	log.Debugf("Profile %s not found", profileName)
	return nil, fmt.Errorf("Error: profile name %s not found", profileName)
}

// GetCredentials returns credentials for a profile in the credentials file if it exists and
// is not expired, or nil otherwise
func (c *AWSConfigProvider) GetCredentials(profileName string) (*CredentialsValue, error) {
	log.Debugf("Looking up existing credentials for %s", profileName)
	if section := c.awsCredentialsFile.Section(profileName); section != nil {
		log.Debug("Found credentials")
		var credentials = &CredentialsValue{}
		if err := section.MapTo(credentials); err != nil {
			log.WithError(err)
			return nil, err
		}
		return credentials, nil
	}
	return nil, nil
}

// NewAWSConfigProvider returns a pointer to a new instance of the AWSConfigProvider
func NewAWSConfigProvider(configPath string, credentialsPath string) (*AWSConfigProvider, error) {
	log.Debug("Creating AWSConfigProvider")
	configPath, err := homedir.Expand(configPath)
	if err != nil {
		log.WithError(err)
		return nil, err
	}
	credentialsPath, err = homedir.Expand(credentialsPath)
	if err != nil {
		log.WithError(err)
		return nil, err
	}
	configFile, err := ini.Load(configPath)
	if err != nil {
		log.WithError(err)
		return nil, err
	}
	credsFile, err := ini.Load(credentialsPath)
	if err != nil {
		log.WithError(err)
		return nil, err
	}
	var configProvider = new(AWSConfigProvider)
	configProvider.configPath = configPath
	configProvider.credentialsPath = credentialsPath
	configProvider.awsConfigFile = configFile
	configProvider.awsCredentialsFile = credsFile
	log.Debugf("AWSConfigProvider created with config file (%s) and creds file (%s)", configPath, credentialsPath)
	return configProvider, nil
}

// SetCredentials stores the provided credentials in the credentials file
func (c *AWSConfigProvider) SetCredentials(profileName string, credentials *CredentialsValue) error {
	section := c.awsCredentialsFile.Section(profileName)
	if err := section.ReflectFrom(credentials); err != nil {
		return err
	}
	return c.awsCredentialsFile.SaveTo(c.credentialsPath)
}

// STSCredentialsProvider fetches credentials from the AWS STS Service
type STSCredentialsProvider struct{}

// AssumeRole calls sts:AssumeRole and returns temporary credentials
func (s *STSCredentialsProvider) AssumeRole(options AssumeRoleOptions) (*CredentialsValue, error) {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: options.SourceProfile,
	}))
	creds := stscreds.NewCredentials(awsSession, options.RoleARN, func(p *stscreds.AssumeRoleProvider) {
		p.Duration = time.Duration(options.SessionDuration) * time.Minute
		if options.MFASerial != "" {
			p.SerialNumber = &options.MFASerial
			p.TokenProvider = stscreds.StdinTokenProvider
		}
		if options.ExternalID != "" {
			p.ExternalID = &options.ExternalID
		}
		if options.RoleSessionName != "" {
			p.RoleSessionName = options.RoleSessionName
		}
	})
	val, err := creds.Get()
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr == credentials.ErrNoValidProvidersFoundInChain {
				log.Errorf("No valid credentials found for source profile %s", options.SourceProfile)
			}
		} else {
			log.Error(err)
		}
		return nil, err
	}
	credsValue := &CredentialsValue{
		AccessKeyID:       val.AccessKeyID,
		SecretAccessKey:   val.SecretAccessKey,
		SessionToken:      val.SessionToken,
		SessionExpiration: time.Now().Local().Add(time.Duration(options.SessionDuration) * time.Minute),
	}
	return credsValue, nil
}

// CredentialsClient manages locally stored data and fetching fresh credentials
type CredentialsClient struct {
	ConfigProvider      ConfigProvider
	CredentialsProvider CredentialsProvider
}

// NewCredentialsClient creates a new credentials client that can assume role and fetch temporary credentials
func NewCredentialsClient(configPath string, credentialsPath string) (*CredentialsClient, error) {
	log.Debug("Creating CredentialsClient")
	configprovider, err := NewAWSConfigProvider(configPath, credentialsPath)
	if err != nil {
		return nil, err
	}
	var credentialsClient = new(CredentialsClient)
	credentialsClient.ConfigProvider = configprovider
	credentialsClient.CredentialsProvider = new(STSCredentialsProvider)
	log.Debug("CredentialsClient created")
	return credentialsClient, nil
}

// GetCredentials retrieves credentials from the credentials file. If they are not valid
// or not present, fresh credentials are fetched from the STS service
func (c *CredentialsClient) GetCredentials(options AssumeRoleOptions) (*CredentialsValue, error) {
	credentials, err := c.ConfigProvider.GetCredentials(options.ProfileName)
	if err != nil {
		return nil, err
	}
	if isValid(credentials) {
		return credentials, nil
	}
	if credentials, err = c.CredentialsProvider.AssumeRole(options); err == nil {
		log.Debug("Got credentials from STS")
		c.ConfigProvider.SetCredentials(options.ProfileName, credentials)
		return credentials, nil
	}
	log.Error(err)
	return nil, err
}

func isValid(credentials *CredentialsValue) bool {
	if credentials == nil {
		log.Debug("No credentials present")
		return false
	}
	if isExpired(credentials) {
		log.Debug("Credentials are expired")
		return false
	}
	return true
}

func isExpired(credentials *CredentialsValue) bool {
	duration := time.Until(credentials.SessionExpiration)
	return duration <= 0
}

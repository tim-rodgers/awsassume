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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-ini/ini"
	homedir "github.com/mitchellh/go-homedir"
)

// A Profile contains the properties for a profile
// stored in ~/.aws/config
type Profile struct {
	SourceProfile   string `ini:"source_profile"`
	RoleArn         string `ini:"role_arn"`
	MfaSerial       string `ini:"mfa_serial"`
	ExternalID      string `ini:"external_id"`
	Region          string `ini:"region"`
	RoleSessionName string `ini:"role_session_name"`
}

// A Value is the credentials value for a particular set of credentials
type Value struct {
	AccessKeyID     string    `ini:"aws_access_key_id"`
	SecretAccessKey string    `ini:"aws_secret_access_key"`
	SessionToken    string    `ini:"aws_session_token"`
	ExpiresAt       time.Time `ini:"awsassume_expires_at"`
	Region          string    `ini:"awsassume_region"`
}

// CredentialProvider is used to retrieve credentials from file or STS API
type CredentialProvider struct {
	ConfigFile    string
	CredsFile     string
	ProfileName   string
	SourceProfile string
	Duration      int
	Region        string
}

// Retrieve retrieves a set of credentials for the specified profile.
// This might be credentials stored in the shared credentials file if valid, or
// retrieved from the STS API
func (c CredentialProvider) Retrieve() (*Value, error) {
	var val = new(Value)
	val = c.CredentialsFromFile()
	if val != nil {
		fmt.Println("Using credentials from file")
		return val, nil
	}
	fmt.Println("Retrieving credentials from AWS STS")
	val, err := c.AssumeRole()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving credentials: %s", err)
	}
	writeCredentials(c.CredsFile, c.ProfileName, val)
	return val, nil
}

// CredentialsFromFile fetches a set of credentials from an AWS shared credentials file
func (c CredentialProvider) CredentialsFromFile() *Value {
	credFile, err := loadIniFile(c.CredsFile)
	if err != nil {
		fmt.Println("Failed to open credentials file: ", err)
		os.Exit(1)
	}
	creds, err := credFile.GetSection(c.ProfileName)
	if err != nil {
		fmt.Println("Profile not found in shared credential file")
		return nil
	}
	val := new(Value)
	err = creds.MapTo(val)
	if err != nil {
		fmt.Println("Profile did not match expected format")
	}
	duration := time.Until(val.ExpiresAt)
	if duration.Minutes() < 1 {
		fmt.Println("Stored credentials have expired")
		return nil
	}
	return val
}

// AssumeRole attempts to assume the role of the specified profile
// and return the temporary credentials
func (c CredentialProvider) AssumeRole() (*Value, error) {
	profile, err := c.GetProfile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var source string

	if c.SourceProfile != "" {
		source = c.SourceProfile
	} else if profile.SourceProfile != "" {
		source = profile.SourceProfile
	} else {
		fmt.Println("Error: No source profile provided")
		os.Exit(1)
	}
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Profile: source,
	}))
	creds := stscreds.NewCredentials(awsSession, profile.RoleArn, func(p *stscreds.AssumeRoleProvider) {
		p.Duration = time.Duration(c.Duration) * time.Minute
		if profile.MfaSerial != "" {
			p.SerialNumber = &profile.MfaSerial
			p.TokenProvider = stscreds.StdinTokenProvider
		}
		if profile.ExternalID != "" {
			p.ExternalID = &profile.ExternalID
		}
		if profile.RoleSessionName != "" {
			p.RoleSessionName = profile.RoleSessionName
		}
	})
	awsVal, err := creds.Get()
	if err != nil {
		return nil, err
	}
	val := new(Value)
	val.AccessKeyID = awsVal.AccessKeyID
	val.SecretAccessKey = awsVal.SecretAccessKey
	val.SessionToken = awsVal.SessionToken
	val.ExpiresAt = time.Now().Local().Add(time.Duration(c.Duration) * time.Minute)
	if c.Region != "" {
		val.Region = c.Region
	} else if profile.Region != "" {
		val.Region = profile.Region
	}
	return val, nil
}

// GetProfile retrieves data about the profile from the AWS CLI config file
func (c CredentialProvider) GetProfile() (*Profile, error) {
	cfg, err := loadIniFile(c.ConfigFile)
	if err != nil {
		fmt.Println("Failed to open config file: ", err)
		os.Exit(1)
	}
	var sectionName string
	if c.ProfileName == "default" {
		sectionName = c.ProfileName
	} else {
		sectionName = fmt.Sprintf("profile %s", c.ProfileName)
	}
	sections := cfg.Sections()
	profile := new(Profile)
	for i := range sections {
		if sections[i].Name() == sectionName {
			sections[i].MapTo(profile)
			return profile, nil
		}
	}
	return nil, errors.New("Profile not found in config file")
}

func loadIniFile(filePath string) (*ini.File, error) {
	path, _ := homedir.Expand(filePath)
	cfg, err := ini.Load(path)
	return cfg, err
}

func writeCredentials(awsCredsPath string, profileName string, val *Value) {
	path, _ := homedir.Expand(awsCredsPath)
	credFile, err := loadIniFile(path)
	if err != nil {
		fmt.Printf("Could not open %s: %v\n", awsCredsPath, err)
		return
	}
	creds := credFile.Section(profileName)
	err = creds.ReflectFrom(val)
	if err != nil {
		fmt.Println("Error reflecting credentials: ", err)
	}
	err = credFile.SaveTo(path)
	if err != nil {
		fmt.Println("Error writing back credentials: ", err)
	}
}

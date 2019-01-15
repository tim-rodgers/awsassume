A tool for running commands with temporary AWS credentials. awsassume makes working with the AWS AssumeRole API easier and more convenient.

# Features

* Run single commands or start a new shell session with profile configuration set in environment
* Supports MFA tokens and ExternalID field
* Stores temporary credentials in the `~/.aws/credentials file` with expiration time
* Supports AWS CLI configuration env vars

# Configuration

awsassume uses your AWS CLI `~/.aws/config` and `~/.aws/credentials` files to retrieve and store temporary credentials from the AWS STS service and run commands.

Example `~/.aws/config` file:

```
[default]
region = us-east-1

[profile prod]
region=eu-west-1
source_profile=default
role_arn=arn:aws:iam::123456789012:role/RoleName
mfa_serial=arn:aws:iam::456789101112:mfa/user
```

Example `~/.aws/credentials` file:

```
[default]
aws_access_key_id=AKIAIOSFODNN7EXAMPLE
aws_secret_access_key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

# Getting started

awsassume has two main commands `run` and `shell`.

The `run` command takes a command as input and will assume the role in the shell you specify (defaults to `$SHELL`):
```
awsassume run --profile prod aws sts get-caller-identity
```

If you want to run several commands you can start a new shell session with credentials set using the `shell command:
```
awsassume shell --profile prod --duration 60
```
Again, the shell launched is sourced from the `$SHELL` env var. For both commands, the shell to use can be set manually with the `--command` flag.

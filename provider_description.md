AWS provider for the Celerity Deploy Engine including resources, data sources, links and custom variable types for interacting with AWS services.

## AWS Authentication and Configuration

The AWS Provider derives authentication and configuration from multiple sources, which are applied in the following order of precedence (as per the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html)):

1. Provider configuration (static credentials, `accessKeyId`, `secretAccessKey` etc.)
1. Environment variables
3. Shared credentials files
4. Shared configuration files
5. Container credentials
6. EC2 instance profile credentials

Provider configuration can be used to customise aspects of multiple sources for credentials and configuration but the order of precedence from the AWS SDK for Go v2 will still apply.

### Provider configuration

> [!warning]
> You should avoid hardcoding credentials in the provider configuration if you use version control to store your deploy configuration files.

Credentials can be provided by using the `accessKeyId` and `secretAccessKey` fields in the provider configuration, optionally with a `sessionToken` field for temporary credentials.

Example usage in a `celerity.deploy.jsonc` file:

```javascript
{
    "providers": {
        "aws": {
            "region": "eu-west-1",
            "accessKeyId": "my-access-key-id",
            "secretAccessKey": "my-secret-key",
            "sessionToken": "my-session-token" // optional
        }
    }
}
```

Additional options can be used to configure authorization such as `profile`, `sharedConfigFiles` and `sharedCredentialsFiles`.

### Environment variables

Credentials can be provided by using the `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_SESSION_TOKEN` environment variables. `AWS_SESSION_TOKEN` is optional.
The region can also be set using the `AWS_REGION` environment variable.

Example usage:

```javascript
{
    "providers": {
        "aws": {}
    }
}
```

```bash
$ export AWS_ACCESS_KEY_ID=my-access-key-id
$ export AWS_SECRET_ACCESS_KEY=my-secret-key
$ export AWS_SESSION_TOKEN=my-session-token # optional
$ export AWS_REGION=eu-west-1
$ celerity stage-changes
```

Additional environment variables can be used to configure authorization such as `AWS_PROFILE`, `AWS_CONFIG_FILE` and `AWS_SHARED_CREDENTIALS_FILE`.

### Shared Credentials and Configuration Files

The AWS provider can be configured to source credentials and configuration from [shared credentials and configuration files](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html). These files are located at `$HOME/.aws/credentials` and `$HOME/.aws/config` by default on Linux and macOS. On Windows, the files are located at `%USERPROFILE%\.aws\credentials` and `%USERPROFILE%\.aws\config`.

If a named profile is not specified, the `default` profile will be used. You can use the `profile` config field or the `AWS_PROFILE` environment variable to specify a different profile.

The locations of the shared credentials and configuration files can be configured using the `shared_credentials_file` and `shared_config_files` config fields or the `AWS_SHARED_CREDENTIALS_FILE` and `AWS_CONFIG_FILE` environment variables.

Example usage in a `celerity.deploy.jsonc` file:

```javascript
{
    "providers": {
        "aws": {
            "sharedCredentialsFiles": "/path/to/creds,/path/to/other-creds",
            "sharedConfigFiles": "/path/to/config,/path/to/other-config",
            "profile": "my-profile"
        }
    }
}
```

### Container Credentials

If you're running the Celerity Deploy Engine in a container on a service such as ECS or CodeBuild and have a configured [IAM Task Role](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html), the AWS provider can use the container's Task Role. This is made possible by the use of the `AWS_CONTAINER_CREDENTIALS_RELATIVE_URI` and `AWS_CONTAINER_CREDENTIALS_FULL_URI` environment variables being automatically set by the container runtime service or set manually in special use cases.

If you're running the Celerity Deploy Engine on EKS and have configured [IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), the AWS provider can use the pod's role. This is made possible by the `AWS_ROLE_ARN` and `AWS_WEB_IDENTITY_TOKEN_FILE` environment variables that are automatically set by the Kubernetes service or manually set in special use cases.

### EC2 Instance Profile Credentials

When the Celerity Deploy Engine is running on an EC2 with an IAM Instance Profile set, the AWS provider can source credentials from the [EC2 Instance Metadata Service](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html). IMDS v1 and v2 are both supported.

A custom endpoint for the metadata service can be configured with the `ec2_metadata_service_endpoint` config field or the `AWS_EC2_METADATA_SERVICE_ENDPOINT` environment variable.

### Assuming an IAM Role

When the `assumeRole.roleArn` config field is set, the AWS provider will attempt to assume the provided role using the supplied credentials.

Example usage in a `celerity.deploy.jsonc` file:

```javascript
{
    "providers": {
        "aws": {
            "assumeRole.roleArn": "arn:aws:iam::123456789012:role/my-role",
            "assumeRole.sessionName": "my-session-name",
            "assumeRole.externalId": "my-external-id"
        }
    }
}
```

### Assuming an IAM Role using a Web Identity

When the `assumeRoleWithWebIdentity.roleArn` config field is set and a web identity token is provided, the AWS provider will attempt to assume the provided role using the supplied credentials.

Example usage in a `celerity.deploy.jsonc` file:

```javascript
{
    "providers": {
        "aws": {
            "assumeRoleWithWebIdentity.roleArn": "arn:aws:iam::123456789012:role/my-role",
            "assumeRoleWithWebIdentity.sessionName": "my-session-name",
            "assumeRoleWithWebIdentity.webIdentityTokenFile": "/path/to/token-file"
        }
    }
}
```

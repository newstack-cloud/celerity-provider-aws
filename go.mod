module github.com/newstack-cloud/bluelink-provider-aws

go 1.26

toolchain go1.26.1

require (
	github.com/aws/aws-sdk-go-v2 v1.42.0
	github.com/aws/aws-sdk-go-v2/config v1.29.15
	github.com/aws/aws-sdk-go-v2/credentials v1.17.68
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.53.5
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.234.0
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.46.6
	github.com/aws/aws-sdk-go-v2/service/iam v1.42.2
	github.com/aws/aws-sdk-go-v2/service/lambda v1.71.2
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.26.7
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.3
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.20
	github.com/aws/smithy-go v1.27.1
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/newstack-cloud/bluelink/libs/blueprint v0.42.1
	github.com/newstack-cloud/bluelink/libs/plugin-framework v0.5.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/coreos/go-json v0.0.0-20231102161613-e49c8866685a // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/newstack-cloud/bluelink/libs/common v0.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/tailscale/hujson v0.0.0-20250226034555-ec1d1c113d33 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

//go:build unit

package lambda

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func createBaseTestFunctionConfig(
	functionName string,
	runtime types.Runtime,
	handler string,
	role string,
) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String(functionName),
			FunctionArn:  aws.String(fmt.Sprintf("arn:aws:lambda:us-east-1:123456789012:function:%s", functionName)),
			Runtime:      runtime,
			Handler:      aws.String(handler),
			Role:         aws.String(role),
			Architectures: []types.Architecture{
				types.ArchitectureX8664,
			},
		},
		Code: &types.FunctionCodeLocation{
			Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
		},
	}
}

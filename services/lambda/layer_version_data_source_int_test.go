//go:build integration

package lambda_test

import (
	"archive/zip"
	"bytes"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/integration"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

type LayerVersionDataSourceIntegrationSuite struct {
	suite.Suite
}

// Test_layer_version_data_source publishes a real Lambda layer version (with an
// inline zip) and fetches it back through the layer version data
// source (GetLayerVersion), by layer name and version number.
func (s *LayerVersionDataSourceIntegrationSuite) Test_layer_version_data_source() {
	h := integration.Setup(s.T())

	layerName := h.Name("ds-layer")
	lambdaClient := lambda.NewFromConfig(h.AWSConfig)
	publishOut, err := lambdaClient.PublishLayerVersion(h.Ctx, &lambda.PublishLayerVersionInput{
		LayerName:          aws.String(layerName),
		Content:            &lambdatypes.LayerVersionContentInput{ZipFile: s.layerZip()},
		CompatibleRuntimes: []lambdatypes.Runtime{lambdatypes.RuntimeNodejs20x},
	})
	s.Require().NoError(err)
	version := publishOut.Version
	s.T().Cleanup(func() {
		_, _ = lambdaClient.DeleteLayerVersion(h.Ctx, &lambda.DeleteLayerVersionInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})
	})

	data := h.Fetch(s.T(), "aws/lambda/layerVersion", integration.Filters(
		integration.EqFilter("layerName", layerName),
		integration.EqFilter("versionNumber", strconv.FormatInt(version, 10)),
	))
	s.Require().Equal(aws.ToString(publishOut.LayerArn), core.StringValue(data["arn"]))
	s.Require().Equal(int(version), core.IntValue(data["version"]))
}

func (s *LayerVersionDataSourceIntegrationSuite) layerZip() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("nodejs/node_modules/placeholder.txt")
	s.Require().NoError(err)
	_, err = w.Write([]byte("bluelink integration test layer"))
	s.Require().NoError(err)
	s.Require().NoError(zw.Close())
	return buf.Bytes()
}

func TestLayerVersionDataSourceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(LayerVersionDataSourceIntegrationSuite))
}

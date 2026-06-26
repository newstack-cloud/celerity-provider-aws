//go:build integration

package e2e

import (
	"fmt"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/require"
)

// TestLambdaDataSources covers the Lambda function-based data sources. Each subtest runs
// in parallel with its own harness (a unique name prefix and isolated state container) —
// required here because they all deploy the same function_resource.blueprint (which
// creates "<prefix>-fn"), so a shared prefix would collide. Concurrency against AWS is
// bounded by the harness operation gate.
func TestLambdaDataSources(t *testing.T) {
	t.Parallel()

	t.Run("function", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "function_resource.blueprint", nil)

		changeSet := h.Stage(t, "function_datasource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s-fn", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(ResolvedExport(t, changeSet, "functionArn")))
	})

	t.Run("alias", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "function_resource.blueprint", nil)
		version := h.PublishAlias(t, h.NamePrefix+"-fn", "live")

		changeSet := h.Stage(t, "alias_datasource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s-fn:live", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(ResolvedExport(t, changeSet, "aliasArn")))
		require.Equal(t, version, core.StringValue(ResolvedExport(t, changeSet, "aliasFunctionVersion")))
	})

	t.Run("function url", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		h.Deploy(t, "function_resource.blueprint", nil)
		h.CreateFunctionURL(t, h.NamePrefix+"-fn")

		changeSet := h.Stage(t, "functionurl_datasource.blueprint", nil)
		expectedARN := fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s-fn", h.Region, h.Account.AccountID, h.NamePrefix)
		require.Equal(t, expectedARN, core.StringValue(ResolvedExport(t, changeSet, "functionArn")))
		require.Equal(t, "NONE", core.StringValue(ResolvedExport(t, changeSet, "authType")))
	})

	t.Run("code signing config", func(t *testing.T) {
		t.Parallel()
		h := Setup(t)
		profileVersionARN := h.SigningProfileVersionARN(t, "csc_ds_profile")
		cscARN := h.CreateCodeSigningConfig(t, profileVersionARN)

		changeSet := h.Stage(t, "codesigningconfig_datasource.blueprint", map[string]*core.ScalarValue{
			"cscArn": core.ScalarFromString(cscARN),
		})
		require.Equal(t, cscARN, core.StringValue(ResolvedExport(t, changeSet, "cscArn")))
	})
}

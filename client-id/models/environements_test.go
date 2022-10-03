package models_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/SKF/go-enlight-middleware/client-id/models"
)

func TestEnvironments_ContainsValid(t *testing.T) {
	envs := models.Environments{models.Sandbox}

	require.True(t, envs.Contains(models.Sandbox))
}

func TestEnvironments_ContainsMissing(t *testing.T) {
	envs := models.Environments{models.Sandbox}

	require.False(t, envs.Contains(models.Test))
}

func TestEnvironments_ContainsEmpty(t *testing.T) {
	envs := models.Environments{}

	require.True(t, envs.Contains(models.Test))
}

func TestEnvironments_ContainsInvalidEmpty(t *testing.T) {
	envs := models.Environments{}

	require.False(t, envs.Contains(models.Environment("FOOBAR")))
}

func TestEnvironmentMask_DisjointSingle(t *testing.T) {
	onlySandbox := models.Environments{models.Sandbox}.Mask()
	onlyTest := models.Environments{models.Test}.Mask()

	require.True(t, onlySandbox.Disjoint(onlyTest))
}

func TestEnvironmentMask_DisjointMultiple(t *testing.T) {
	sandboxTest := models.Environments{models.Sandbox, models.Test}.Mask()
	stagingProd := models.Environments{models.Staging, models.Prod}.Mask()

	require.True(t, sandboxTest.Disjoint(stagingProd))
}

func TestEnvironmentMask_DisjointSubset(t *testing.T) {
	sandboxTest := models.Environments{models.Sandbox, models.Test}.Mask()
	onlyTest := models.Environments{models.Test}.Mask()

	require.False(t, sandboxTest.Disjoint(onlyTest))
}

func TestEnvironmentMask_DisjointAll(t *testing.T) {
	all := models.Environments{}.Mask()
	onlyProd := models.Environments{models.Test}.Mask()

	require.False(t, all.Disjoint(onlyProd))
}

func TestEnvironmentMask_DisjointInvalid(t *testing.T) {
	all := models.Environments{}.Mask()
	invalid := models.Environments{"FOOBAR"}.Mask()

	require.True(t, all.Disjoint(invalid))
}

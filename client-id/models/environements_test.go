package models_test

import (
	"testing"

	"github.com/SKF/go-enlight-middleware/client-id/models"
	"github.com/stretchr/testify/require"
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

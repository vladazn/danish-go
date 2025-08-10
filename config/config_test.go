package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {

	t.Setenv("FIREBASE_PROJECT_ID", "123")

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.Equal(t, "123", cfg.FirebaseConfig.ProjectId)
}

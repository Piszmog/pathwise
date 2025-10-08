package utils_test

import (
	"testing"

	"github.com/Piszmog/pathwise/internal/ui/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!@#$%^&*()",
		},
		{
			name:     "unicode password",
			password: "пароль123",
		},
		{
			name:     "long password",
			password: "this_is_a_very_long_password_that_should_still_work_correctly_123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hash, err := utils.HashPassword([]byte(tt.password), 4)

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, string(hash))

			// Verify the hash can be used to check the password
			err = utils.CheckPasswordHash(hash, []byte(tt.password))
			assert.NoError(t, err)
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	t.Parallel()
	password := "testpassword123"
	hash, err := utils.HashPassword([]byte(password), 4)
	require.NoError(t, err)

	tests := []struct {
		name        string
		hash        []byte
		password    string
		expectError bool
	}{
		{
			name:        "correct password",
			hash:        hash,
			password:    password,
			expectError: false,
		},
		{
			name:        "incorrect password",
			hash:        hash,
			password:    "wrongpassword",
			expectError: true,
		},
		{
			name:        "empty password against valid hash",
			hash:        hash,
			password:    "",
			expectError: true,
		},
		{
			name:        "password against empty hash",
			hash:        []byte(""),
			password:    password,
			expectError: true,
		},
		{
			name:        "password against invalid hash",
			hash:        []byte("invalid_hash"),
			password:    password,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := utils.CheckPasswordHash(tt.hash, []byte(tt.password))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	t.Parallel()
	password := "consistencytest"

	hash1, err1 := utils.HashPassword([]byte(password), 4)
	require.NoError(t, err1)

	hash2, err2 := utils.HashPassword([]byte(password), 4)
	require.NoError(t, err2)

	// Hashes should be different (bcrypt includes salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should validate the same password
	assert.NoError(t, utils.CheckPasswordHash(hash1, []byte(password)))
	assert.NoError(t, utils.CheckPasswordHash(hash2, []byte(password)))
}

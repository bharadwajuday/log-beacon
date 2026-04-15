package auth

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Test check
	assert.True(t, CheckPasswordHash(password, hash))
	assert.False(t, CheckPasswordHash("wrongpassword", hash))
}

func TestGenerateAndValidateJWT(t *testing.T) {
	username := "testuser"
	token, err := GenerateJWT(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ValidateJWT(token)
	assert.NoError(t, err)
	assert.Equal(t, username, claims.Username)

	_, err = ValidateJWT("invalid.token.here")
	assert.Error(t, err)
}

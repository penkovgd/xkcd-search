package aaa

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yadro.com/course/api/core"
)

const (
	testAdminUser     = "test_admin"
	testAdminPassword = "test_password_123"
)

func TestLogin_Success(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	// Act
	aaa, err := New(tokenTTL, log)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, aaa)

	token, err := aaa.Login(testAdminUser, testAdminPassword)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	err = aaa.Verify(token)
	assert.NoError(t, err)
}

func TestNew_MissingEnv(t *testing.T) {
	// Act
	aaa, err := New(1*time.Hour, slog.Default())

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "admin user")
	assert.Equal(t, AAA{}, aaa)
}

func TestLogin_InvalidUsername(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	aaa, err := New(tokenTTL, log)
	require.NoError(t, err)

	// Act
	token, err := aaa.Login("wrong_user", testAdminPassword)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrInvalidCredentials)
	assert.Empty(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	aaa, err := New(tokenTTL, log)
	require.NoError(t, err)

	// Act
	token, err := aaa.Login(testAdminUser, "wrong_password")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrInvalidCredentials)
	assert.Empty(t, token)
}

func TestVerify_ValidToken(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	aaa, err := New(tokenTTL, log)
	require.NoError(t, err)

	token, err := aaa.Login(testAdminUser, testAdminPassword)
	require.NoError(t, err)

	// Act
	err = aaa.Verify(token)

	// Assert
	assert.NoError(t, err)
}

func TestVerify_InvalidTokenFormat(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	aaa, err := New(tokenTTL, log)
	require.NoError(t, err)

	invalidTokens := []string{
		"",
		"not.a.token",
		"invalid.token.signature",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", // Токен с другим ключом
	}

	for _, invalidToken := range invalidTokens {
		t.Run(fmt.Sprintf("token_%s", invalidToken), func(t *testing.T) {
			// Act
			err := aaa.Verify(invalidToken)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, core.ErrInvalidToken)
		})
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	aaa := AAA{
		users:    map[string]string{testAdminUser: testAdminPassword},
		tokenTTL: -1 * time.Hour,
		log:      log,
	}

	token, err := aaa.generateToken()
	require.NoError(t, err)

	// Act
	err = aaa.Verify(token)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrInvalidToken)
}

func TestVerify_TokenWithWrongSubject(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "not_superuser",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
	})

	wrongToken, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err)

	aaa := AAA{
		users:    map[string]string{testAdminUser: testAdminPassword},
		tokenTTL: tokenTTL,
		log:      log,
	}

	// Act
	err = aaa.Verify(wrongToken)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, core.ErrInvalidToken)
}

func TestGenerateToken_Success(t *testing.T) {
	// Arrange
	t.Setenv("ADMIN_USER", testAdminUser)
	t.Setenv("ADMIN_PASSWORD", testAdminPassword)

	log := slog.Default()
	tokenTTL := 1 * time.Hour

	aaa := AAA{
		users:    map[string]string{testAdminUser: testAdminPassword},
		tokenTTL: tokenTTL,
		log:      log,
	}

	// Act
	token, err := aaa.generateToken()

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Check subject
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, adminRole, claims["sub"])

	// Check expiration
	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	expTime := time.Unix(int64(exp), 0)
	assert.True(t, expTime.After(time.Now()))
}

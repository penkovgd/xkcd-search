package aaa

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"yadro.com/course/api/core"
)

const secretKey = "something secret here" // token sign key
const adminRole = "superuser"             // token subject

// Authentication, Authorization, Accounting
type AAA struct {
	users    map[string]string
	tokenTTL time.Duration
	log      *slog.Logger
}

func New(tokenTTL time.Duration, log *slog.Logger) (AAA, error) {
	const adminUser = "ADMIN_USER"
	const adminPass = "ADMIN_PASSWORD"
	user, ok := os.LookupEnv(adminUser)
	if !ok {
		return AAA{}, fmt.Errorf("could not get admin user from enviroment")
	}
	password, ok := os.LookupEnv(adminPass)
	if !ok {
		return AAA{}, fmt.Errorf("could not get admin password from enviroment")
	}

	return AAA{
		users:    map[string]string{user: password},
		tokenTTL: tokenTTL,
		log:      log,
	}, nil
}

func (a AAA) Login(name, password string) (string, error) {
	pass, ok := a.users[name]
	if !ok {
		return "", core.ErrInvalidCredentials
	}
	if pass != password {
		return "", core.ErrInvalidCredentials
	}

	token, err := a.generateToken()
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	return token, nil
}

func (a AAA) Verify(tokenString string) error {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		a.log.Error("cannot parse token", "error", err)
		return core.ErrInvalidToken
	}
	if !token.Valid {
		a.log.Error("token is invalid")
		return core.ErrInvalidToken
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		a.log.Error("no subject", "error", err)
		return core.ErrInvalidToken
	}
	if subject != adminRole {
		a.log.Error("not admin", "subject", subject)
		return core.ErrInvalidToken
	}
	return nil
}

func (a AAA) generateToken() (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   adminRole,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenTTL)),
	})
	signed, err := t.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return signed, nil
}

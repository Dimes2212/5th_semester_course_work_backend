package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
	"github.com/golang-jwt/jwt/v5"
)

func makeAuthResult(foundUser user.User, tokens TokenPair) AuthResult {
	return AuthResult{
		User: AuthUser{
			ID:        foundUser.ID,
			FullName:  foundUser.FullName,
			Email:     foundUser.Email,
			Role:      foundUser.Role,
			AvatarKey: foundUser.AvatarKey,
			CreatedAt: foundUser.CreatedAt,
			UpdatedAt: foundUser.UpdatedAt,
		},
		Tokens: tokens,
	}
}

func (s *Service) generateAccessToken(foundUser user.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.accessTokenTTL)

	claims := AccessTokenClaims{
		UserID: foundUser.ID,
		Role:   foundUser.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.jwtIssuer,
			Subject:   foundUser.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (s *Service) generateRefreshToken() (
	token string,
	tokenHash string,
	expiresAt time.Time,
	err error,
) {
	tokenBytes := make([]byte, 32)
	if _, err = rand.Read(tokenBytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	token = base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash = hashRefreshToken(token)
	expiresAt = time.Now().UTC().Add(s.refreshTokenTTL)

	return token, tokenHash, expiresAt, nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

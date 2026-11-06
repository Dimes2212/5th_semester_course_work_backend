package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Dimes2212/5th_semester_course_work_backend/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = 12

type Service struct {
	repository      *Repository
	jwtSecret       string
	jwtIssuer       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewService(
	repository *Repository,
	jwtSecret string,
	jwtIssuer string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		repository:      repository,
		jwtSecret:       jwtSecret,
		jwtIssuer:       jwtIssuer,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	input, err := validateRegisterInput(input)
	if err != nil {
		return AuthResult{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), passwordHashCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	createdUser, err := s.repository.CreateUser(ctx, input.FullName, input.Email, string(passwordHash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return AuthResult{}, ErrEmailAlreadyExists
		}

		return AuthResult{}, fmt.Errorf("create user: %w", err)
	}

	return s.issueTokens(ctx, createdUser, input.DeviceName)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	input, err := normalizeLoginInput(input)
	if err != nil {
		return AuthResult{}, err
	}

	foundUser, err := s.repository.FindUserByEmail(ctx, input.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, fmt.Errorf("find user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, foundUser, input.DeviceName)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (AuthResult, error) {
	if refreshToken == "" {
		return AuthResult{}, ErrRefreshTokenRequired
	}

	oldTokenHash := hashRefreshToken(refreshToken)
	newRefreshToken, newRefreshTokenHash, refreshExpiresAt, err := s.generateRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshUser, err := s.repository.RotateRefreshSession(
		ctx,
		oldTokenHash,
		newRefreshTokenHash,
		refreshExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthResult{}, ErrInvalidRefreshToken
	}
	if err != nil {
		return AuthResult{}, fmt.Errorf("rotate refresh session: %w", err)
	}

	accessToken, accessExpiresAt, err := s.generateAccessToken(refreshUser)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate access token: %w", err)
	}

	return makeAuthResult(refreshUser, TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          newRefreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}), nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return ErrRefreshTokenRequired
	}

	if err := s.repository.RevokeRefreshSession(
		ctx,
		hashRefreshToken(refreshToken),
	); err != nil {
		return fmt.Errorf("revoke refresh session: %w", err)
	}

	return nil
}

func validateRegisterInput(input RegisterInput) (RegisterInput, error) {
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.FullName == "" {
		return RegisterInput{}, ErrFullNameRequired
	}
	if utf8.RuneCountInString(input.FullName) > 100 {
		return RegisterInput{}, ErrFullNameTooLong
	}
	if len(input.Email) > 255 {
		return RegisterInput{}, ErrEmailTooLong
	}

	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil || parsedEmail.Address != input.Email {
		return RegisterInput{}, ErrInvalidEmail
	}
	if utf8.RuneCountInString(input.Password) < 8 {
		return RegisterInput{}, ErrPasswordTooShort
	}
	if len([]byte(input.Password)) > 72 {
		return RegisterInput{}, ErrPasswordTooLong
	}
	if input.Password != input.PasswordConfirmation {
		return RegisterInput{}, ErrPasswordsDoNotMatch
	}

	input.DeviceName, err = normalizeDeviceName(input.DeviceName)
	if err != nil {
		return RegisterInput{}, err
	}

	return input, nil
}

func normalizeLoginInput(input LoginInput) (LoginInput, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	deviceName, err := normalizeDeviceName(input.DeviceName)
	if err != nil {
		return LoginInput{}, err
	}

	input.DeviceName = deviceName
	return input, nil
}

func normalizeDeviceName(deviceName *string) (*string, error) {
	if deviceName == nil {
		return nil, nil
	}

	normalized := strings.TrimSpace(*deviceName)
	if utf8.RuneCountInString(normalized) > 100 {
		return nil, ErrDeviceNameTooLong
	}
	if normalized == "" {
		return nil, nil
	}

	return &normalized, nil
}

func (s *Service) issueTokens(
	ctx context.Context,
	foundUser user.User,
	deviceName *string,
) (AuthResult, error) {
	accessToken, accessExpiresAt, err := s.generateAccessToken(foundUser)
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, refreshTokenHash, refreshExpiresAt, err := s.generateRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.repository.CreateRefreshSession(ctx, RefreshSession{
		UserID:     foundUser.ID,
		TokenHash:  refreshTokenHash,
		DeviceName: deviceName,
		ExpiresAt:  refreshExpiresAt,
	}); err != nil {
		return AuthResult{}, fmt.Errorf("create refresh session: %w", err)
	}

	return makeAuthResult(foundUser, TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}), nil
}

// функции для миддлвары ----------------------------------------------------------
func (s *Service) ParseAccessToken(accessToken string) (Identity, error) {
	if accessToken == "" {
		return Identity{}, ErrAccessTokenRequired
	}

	claims := &AccessTokenClaims{}

	token, err := jwt.ParseWithClaims(accessToken, claims, func(_ *jwt.Token) (any, error) {
		return []byte(s.jwtSecret), nil
	},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(s.jwtIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil || !token.Valid {
		return Identity{}, ErrInvalidAccessToken
	}

	if claims.UserID == "" {
		return Identity{}, ErrInvalidAccessToken
	}

	if claims.Subject != claims.UserID {
		return Identity{}, ErrInvalidAccessToken
	}

	if claims.Role != user.RoleUser &&
		claims.Role != user.RoleAdmin {
		return Identity{}, ErrInvalidAccessToken
	}

	return Identity{
		UserID: claims.UserID,
		Role:   claims.Role,
	}, nil
}

package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"live-polling-app/backend/internal/models"
	"live-polling-app/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyUsed   = errors.New("email is already registered")
	ErrInvalidInput       = errors.New("invalid input")
)

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret []byte
	jwtIssuer string
	jwtExpiry time.Duration
}

type AuthResult struct {
	Token string
	User  models.User
}

func NewAuthService(users *repository.UserRepository, secret, issuer string, expiry time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(secret),
		jwtIssuer: issuer,
		jwtExpiry: expiry,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (AuthResult, error) {
	email, err := normalizeEmail(email)
	if err != nil || len(password) < 8 {
		return AuthResult{}, ErrInvalidInput
	}
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return AuthResult{}, ErrEmailAlreadyUsed
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return AuthResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}
	user := models.User{
		ID:           bson.NewObjectID(),
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		if mongoDuplicateKey(err) {
			return AuthResult{}, ErrEmailAlreadyUsed
		}
		return AuthResult{}, err
	}
	token, err := s.issueToken(user)
	return AuthResult{Token: token, User: user}, err
}

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthResult, error) {
	normalized, err := normalizeEmail(email)
	if err != nil || password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}
	user, err := s.users.FindByEmail(ctx, normalized)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	token, err := s.issueToken(user)
	return AuthResult{Token: token, User: user}, err
}

func (s *AuthService) ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	}, jwt.WithIssuer(s.jwtIssuer))
	if err != nil || !token.Valid {
		return "", ErrInvalidCredentials
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidCredentials
	}
	subject, ok := claims["sub"].(string)
	if !ok || subject == "" {
		return "", ErrInvalidCredentials
	}
	return subject, nil
}

func (s *AuthService) issueToken(user models.User) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.Hex(),
		Issuer:    s.jwtIssuer,
		ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func normalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 {
		return "", ErrInvalidInput
	}
	return email, nil
}

func mongoDuplicateKey(err error) bool {
	var writeErr mongo.WriteException
	return errors.As(err, &writeErr) && writeErr.HasErrorCode(11000)
}

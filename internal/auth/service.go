package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

// Role represents a user role
type Role string

const (
	RoleCustomer        Role = "customer"
	RoleRestaurantOwner Role = "restaurant_owner"
	RoleCourier         Role = "courier"
	RoleAdmin           Role = "admin"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         Role
	FirstName    string
	LastName     string
	IsSuspended  bool
	CreatedAt    time.Time
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Repository defines the interface for user data access
type Repository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user *User) error
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

// Service handles authentication operations
type Service struct {
	repo       Repository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New creates a new auth service
func New(repo Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, email, password, role, firstName, lastName string) (*User, error) {
	// Check if user already exists
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         Role(role),
		FirstName:    firstName,
		LastName:     lastName,
		IsSuspended:  false,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.IsSuspended {
		return nil, ErrUserSuspended
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Hash the refresh token to look it up
	tokenHash := hashToken(refreshToken)

	// Get refresh token from database
	rt, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if token is revoked
	if rt.RevokedAt != nil {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Get user
	user, err := s.repo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.IsSuspended {
		return nil, ErrUserSuspended
	}

	// Revoke old refresh token
	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout revokes a refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.repo.RevokeRefreshToken(ctx, tokenHash)
}

// LogoutAll revokes all refresh tokens for a user
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllRefreshTokens(ctx, userID)
}

// ValidateAccessToken validates an access token and returns the claims
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrTokenExpired
	}

	return claims, nil
}

// generateAccessToken generates a JWT access token
func (s *Service) generateAccessToken(user *User) (string, error) {
	claims := &Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL).UTC()),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateRefreshToken generates a refresh token
func (s *Service) generateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(s.refreshTTL).UTC()
	tokenHash := hashToken(token)

	if err := s.repo.CreateRefreshToken(ctx, userID, tokenHash, expiresAt); err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return token, nil
}

// hashToken hashes a token for storage
func hashToken(token string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	return string(hash)
}

// VerifyTokenHash verifies a token against its hash
func VerifyTokenHash(token, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
}

package service

import (
	"context"
	"errors"
	"time"

	"github.com/b25/services/content/internal/domain"
	"github.com/b25/services/content/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, input domain.RegisterUserInput) (*domain.User, error)
	Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
	jwtExpiry time.Duration
}

type Claims struct {
	UserID   uuid.UUID       `json:"user_id"`
	Email    string          `json:"email"`
	Username string          `json:"username"`
	Role     domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
	}
}

func (s *authService) Register(ctx context.Context, input domain.RegisterUserInput) (*domain.User, error) {
	// Check if user already exists
	_, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil {
		return nil, domain.ErrUserExists
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Set default role if not specified
	role := input.Role
	if role == "" {
		role = domain.UserRoleAuthor
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Clear password hash before returning
	user.PasswordHash = ""
	return user, nil
}

func (s *authService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "content-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Clear password hash before returning
	user.PasswordHash = ""

	return &domain.AuthResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Clear password hash
	user.PasswordHash = ""
	return user, nil
}

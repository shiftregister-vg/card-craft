package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already taken")
	ErrUsernameTaken      = errors.New("username already taken")
)

type Service struct {
	secretKey []byte
	UserStore *models.UserStore
}

func NewService(secretKey string, userStore *models.UserStore) *Service {
	return &Service{
		secretKey: []byte(secretKey),
		UserStore: userStore,
	}
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a password with its hash
func (s *Service) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates a JWT token and returns the user ID
func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", errors.New("invalid token")
}

// Authenticate authenticates a user with email and password
func (s *Service) Authenticate(user *models.User, password string) error {
	if user == nil {
		return ErrUserNotFound
	}

	if !s.CheckPassword(password, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	return nil
}

// Register handles user registration
func (s *Service) Register(username string, email string, password string) (*models.AuthPayload, error) {
	// Check if email is already taken
	existingUser, err := s.UserStore.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailTaken
	}

	// Check if username is already taken
	existingUser, err = s.UserStore.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameTaken
	}

	// Hash the password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	// Save user to database
	if err := s.UserStore.Create(user); err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

// Login handles user login with either username or email
func (s *Service) Login(identifier string, password string) (*models.AuthPayload, error) {
	var user *models.User
	var err error

	// Try to find user by email first
	user, err = s.UserStore.FindByEmail(identifier)
	if err != nil {
		return nil, err
	}

	// If not found by email, try username
	if user == nil {
		user, err = s.UserStore.FindByUsername(identifier)
		if err != nil {
			return nil, err
		}
	}

	// Authenticate user
	if err := s.Authenticate(user, password); err != nil {
		return nil, err
	}

	// Generate token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

// RefreshToken handles token refresh
func (s *Service) RefreshToken(ctx context.Context) (*models.AuthPayload, error) {
	// Get user from context
	user := GetUserFromContext(ctx)
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Generate new token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

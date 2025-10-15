package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTSecret is the secret key for JWT tokens
var JWTSecret = []byte("video-call-server-secret-key-change-in-production")

// User represents a user in the system
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks if a password matches a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID, username string) (string, error) {
	// Set expiration time to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)
	
	// Create claims
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign token with secret key
	return token.SignedString(JWTSecret)
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	// Parse token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})
	
	// Check for parsing errors
	if err != nil {
		return nil, err
	}
	
	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	
	return claims, nil
}

// Mock user storage (in production, use a database)
var users = make(map[string]*User)

// RegisterUser registers a new user
func RegisterUser(username, email, password string) (*User, error) {
	// Check if user already exists
	for _, user := range users {
		if user.Username == username || user.Email == email {
			return nil, errors.New("user already exists")
		}
	}
	
	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	
	// Create user
	user := &User{
		ID:       generateUserID(),
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
	
	// Store user
	users[user.ID] = user
	
	return user, nil
}

// AuthenticateUser authenticates a user with username/email and password
func AuthenticateUser(identifier, password string) (*User, error) {
	// Find user by username or email
	var user *User
	for _, u := range users {
		if u.Username == identifier || u.Email == identifier {
			user = u
			break
		}
	}
	
	// Check if user exists
	if user == nil {
		return nil, errors.New("user not found")
	}
	
	// Check password
	if !CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid password")
	}
	
	return user, nil
}

// GetUserByID returns a user by ID
func GetUserByID(userID string) (*User, bool) {
	user, exists := users[userID]
	return user, exists
}

// generateUserID generates a simple user ID (in production, use UUID)
func generateUserID() string {
	// In production, use uuid.New().String()
	// For simplicity, we'll use a counter
	idCounter++
	return string(rune(idCounter))
}

var idCounter int
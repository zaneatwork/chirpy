package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   userID.String(),
		})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString,
		&jwt.RegisteredClaims{},

		func(token *jwt.Token) (interface{}, error) {
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
      }
			return []byte(tokenSecret), nil
		})


	if err != nil {
		return uuid.UUID{}, err
	}

  subject := token.Claims.(*jwt.RegisteredClaims).Subject
  id, err := uuid.Parse(subject)
  return id, err
}

func GetBearerToken(headers http.Header) (string, error) {
  authHeader := headers.Get("Authorization")
  if len(authHeader) == 0 {
    return "", errors.New("Failed to get Bearer Token, Authorization header missing.")
  }

  headerParts := strings.Split(authHeader, "Bearer ")
  if len(headerParts) < 2 {
    return "", errors.New("Authorization header does not contain Bearer Token.")
  } 
  bearerToken := headerParts[1]

  return bearerToken, nil
}

func MakeRefreshToken() (string, error) {
  tokenB := make([]byte, 32)
  _, err := rand.Read(tokenB) 
  tokenStr := hex.EncodeToString(tokenB)
  return tokenStr, err
}

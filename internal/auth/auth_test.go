package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckTokenCreation(t *testing.T) {
	userId := uuid.New()

	tests := []struct {
		name        string
		userId      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
		wantErr     bool
	}{
		{
			name:        "Valid uuid and expiration.",
			userId:      userId,
			tokenSecret: "blargus",
			expiresIn:   time.Duration(10000) * time.Second,
			wantErr:     false,
		},
		{
			name:        "Invalid expiration.",
			userId:      userId,
			tokenSecret: "bingus",
			expiresIn:   time.Duration(0),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString, err := MakeJWT(tt.userId, tt.tokenSecret, tt.expiresIn)
			foundId, err := ValidateJWT(tokenString, tt.tokenSecret)

			if (err != nil) && !tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if foundId != tt.userId && !tt.wantErr {
				t.Errorf(
					"Failed to validate JWT, UUID's don't match. Want %v, Found %v",
					tt.userId,
					foundId,
				)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		wantErr bool
	}{
		{
			name: "Valid bearer token.",
			headers: http.Header{
				"Authorization": []string{"Bearer 123456"},
			},
			wantErr: false,
		},
		{
			name:    "Missing authorization header.",
			headers: http.Header{
				"Blargus": []string{"Bearer 123456"},
			},
			wantErr: true,
		},
		{
			name:    "Missing bearer token.",
			headers: http.Header{
				"Authorization": []string{"Blargus 123456"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)

			if (err != nil) && !tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
			}
      if (token != "123456") && !tt.wantErr {
        t.Errorf("GetBearerToken() error, failed to return proper token. got: %v", token)
      }
		})
	}
}

package auth

import (
	"net/http"
	"strings"
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
	t.Logf("hash1 length: %d, content: %s", len(hash1), hash1)
	t.Logf("hash1 length: %d, content: %s", len(hash2), hash2)

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

func TestMakeJWT(t *testing.T) {
	// Prepare some data for testing
	tokenSecret := "correctSectetToken"
	anotherSecret := "anotherSectetToken"
	exampleUserID := uuid.MustParse("47985c9a-4bee-45f5-b786-2cdff9045c9b")

	tests := []struct {
		name            string
		userID          uuid.UUID
		createSecret    string
		validateSecret  string
		expiresIn       time.Duration
		wantMakeErr     bool
		wantValidateErr bool
	}{
		{
			name:            "Valid token creation",
			userID:          exampleUserID,
			createSecret:    tokenSecret,
			validateSecret:  tokenSecret,
			expiresIn:       time.Hour,
			wantMakeErr:     false,
			wantValidateErr: false,
		},
		{
			name:            "Empty secret",
			userID:          exampleUserID,
			createSecret:    tokenSecret,
			validateSecret:  "",
			expiresIn:       time.Hour,
			wantMakeErr:     false,
			wantValidateErr: true,
		},
		{
			name:            "Wrong secret",
			userID:          exampleUserID,
			createSecret:    tokenSecret,
			validateSecret:  anotherSecret,
			expiresIn:       time.Hour,
			wantMakeErr:     false,
			wantValidateErr: true,
		},
		{
			name:            "Expired token",
			userID:          exampleUserID,
			createSecret:    tokenSecret,
			validateSecret:  anotherSecret,
			expiresIn:       -1 * time.Hour,
			wantMakeErr:     false,
			wantValidateErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := MakeJWT(tt.userID, tt.createSecret)
			if (err != nil) != tt.wantMakeErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantMakeErr)
			}
			if token == "" {
				t.Error("Expected non-empty token string")
			}
			parts := strings.Split(token, ".")
			if len(parts) != 3 {
				t.Errorf("Expected JWT to have 3 parts, got %d", len(parts))
			}
			returnedUserID, err := ValidateJWT(token, tt.validateSecret)
			if (err != nil) != tt.wantValidateErr {
				t.Errorf("ValidateJWT() error = %v", err)
			}
			if !tt.wantValidateErr && returnedUserID != tt.userID {
				t.Errorf("Expected returnedUserID `%s` to be the same as input userID `%s", returnedUserID, tt.userID)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	header := make(http.Header)
	header["Authorization"] = []string{"Bearer TOKEN"}

	tests := []struct {
		name        string
		header      string
		bearerToken string
		token       string
		wantErr     bool
	}{
		{
			name:        "Token present",
			header:      "Authorization",
			bearerToken: "Bearer TOKEN",
			token:       "TOKEN",
			wantErr:     false,
		},
		{
			name:        "Extra spaces",
			header:      "Authorization",
			bearerToken: "Bearer TOKEN ",
			token:       "TOKEN",
			wantErr:     false,
		},
		{
			name:        "Wrong header",
			header:      "Authentication",
			bearerToken: "Bearer TOKEN ",
			token:       "TOKEN",
			wantErr:     true,
		},
		{
			name:        "Malformed header",
			header:      "Authorization",
			bearerToken: "Barer TOKEN",
			token:       "TOKEN",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := make(http.Header)
			header[tt.header] = []string{tt.bearerToken}
			token, err := GetBearerToken(header)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && token != tt.token {
				t.Errorf("Expected token to be `%s` but got `%s`", tt.token, token)
			}
		})
	}
}

package tokenPackage

import (
	"errors"
	"go-chat-application/config"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var TokenMap = make(map[string]*jwt.Token)

func AddTokenToMap(token *jwt.Token) {
	id := token.Claims.(jwt.MapClaims)["ID"].(string)
	TokenMap[id] = token
}

// ExtractJWTTokenFromHeader extracts the JWT token from the Authorization header of the HTTP request.
func ExtractJWTTokenFromHeader(r *http.Request) string {
	// Get the Authorization header from the request.
	authHeader := r.Header.Get("Authorization")

	// If the Authorization header is empty, return an empty string.
	if authHeader == "" {
		return ""
	}

	// If the Authorization header is longer than 7 characters and contains a space,
	// it's likely in the format "Bearer <token>".
	if len(authHeader) > 7 && strings.Contains(authHeader, " ") {
		// Split the Authorization header into two parts at the first space.
		parts := strings.SplitN(authHeader, " ", 2)

		// If there are two parts, the second part should be the token.
		if len(parts) == 2 {
			return parts[1]
		}
	}

	// If the Authorization header is not in the expected format, return an empty string.
	return ""
}

// ParseAndValidateJWTToken parses and validates the JWT token string.
func ParseAndValidateJWTToken(tokenString string) (*jwt.Token, error) {
	// If the token string is empty, return an error.
	if tokenString == "" {
		return nil, errors.New("no token provided")
	}

	// Parse the token string.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// If the token's signing method is not HMAC, return an error.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		// If the token's signing method is HMAC, return the JWT secret as the key.
		return []byte(config.ApiCfg.JwtSecret), nil
	})

	// If parsing the token string failed, return the error.
	if err != nil {
		return nil, err
	}

	// If parsing the token string succeeded, return the token.
	return token, nil
}

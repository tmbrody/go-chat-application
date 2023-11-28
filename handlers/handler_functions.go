package handlers

import (
	"go-chat-application/config"
	"go-chat-application/internal/database"
	"go-chat-application/tokenPackage"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func ExtractDBAndToken(r *http.Request) (*jwt.Token, string, *database.MongoDBClient) {
	// Extract the JWT token from the request header
	tokenString := tokenPackage.ExtractJWTTokenFromHeader(r)

	// If the token string is empty, return nil values
	if tokenString == "" {
		return nil, "", nil
	}

	// Parse and validate the JWT token
	token, err := tokenPackage.ParseAndValidateJWTToken(tokenString)
	// If there's an error in parsing or validating, return nil values
	if err != nil {
		return nil, "", nil
	}

	// Loop through the tokens in the TokenMap
	for _, tok := range tokenPackage.TokenMap {
		// If the ID claim of the current token matches the ID claim of the parsed token, assign the current token to the parsed token
		if tok.Claims.(jwt.MapClaims)["ID"] == token.Claims.(jwt.MapClaims)["ID"] {
			token = tok
			break
		}
	}

	// If the Revoked claim of the token is true, return nil values
	if token.Claims.(jwt.MapClaims)["Revoked"] == true {
		return nil, "", nil
	}

	// Extract the MongoDB client from the request context
	ctx := r.Context()
	client, errBool := ctx.Value(config.ApiCfg.DB).(*database.MongoDBClient)
	// If there's an error in type assertion, return nil values
	if !errBool {
		return nil, "", nil
	}

	// Return the token, token string, and MongoDB client
	return token, tokenString, client
}

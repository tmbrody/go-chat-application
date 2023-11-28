package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go-chat-application/config"
	"go-chat-application/internal/database"
	"go-chat-application/tokenPackage"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Define constants for access and refresh token expiration times
const AccessExpiration time.Duration = time.Hour
const RefreshExpiration time.Duration = 7 * (time.Hour * 24)

// CreateUserHandler is a HTTP handler function that creates a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the MongoDB client from the request context
	ctx := r.Context()
	client, _ := ctx.Value(config.ApiCfg.DB).(*database.MongoDBClient)

	// Define a struct to hold the request parameters
	var params struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body into the params struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Create a new user using the provided parameters
	user, err := client.CreateUser(params.Name, params.Email, params.Password)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to create user")
		return
	}

	// Create a map to hold the user data
	userMap := map[string]interface{}{
		"_id":        user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	// Respond with the created user data
	RespondWithJSON(w, http.StatusCreated, userMap)
}

// GetUsersHandler is a HTTP handler function that retrieves all users
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the MongoDB client from the request context
	ctx := r.Context()
	client, _ := ctx.Value(config.ApiCfg.DB).(*database.MongoDBClient)

	// Retrieve all users
	users, err := client.GetAllUsers()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to get users")
		return
	}

	// Create a slice of maps to hold the user data
	var userMap []map[string]interface{}

	// Loop over the users and add their data to the userMap slice
	for _, user := range users {
		userMap = append(userMap,
			map[string]interface{}{
				"_id":        user.ID,
				"name":       user.Name,
				"email":      user.Email,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
		)
	}

	// Respond with the user data
	RespondWithJSON(w, http.StatusOK, userMap)
}

// UpdateUserHandler handles the user update request
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the token and the database client from the request
	token, _, client := ExtractDBAndToken(r)

	// If the token is nil, return immediately
	if token == nil {
		return
	}

	// Get the issuer from the token claims
	issuer := token.Claims.(jwt.MapClaims)["Issuer"].(string)

	// If the issuer is a refresh token, respond with an error
	if issuer == "go-chat-application-refresh" {
		RespondWithError(w, http.StatusUnauthorized,
			"Using JWT refresh token when JWT access token is required")
		return
	}

	// Define the parameters structure
	var params struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body into the parameters structure
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get all users from the database
	users, err := client.GetAllUsers()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to get users")
		return
	}

	// Get the user ID from the token claims
	userID := token.Claims.(jwt.MapClaims)["Subject"].(string)
	userID = strings.TrimPrefix(userID, "ObjectID(\"")
	userID = strings.TrimSuffix(userID, "\")")

	// If the user ID is empty, respond with an error
	if userID == "" {
		RespondWithError(w, http.StatusUnauthorized, "Unable to get user ID from JWT token")
		return
	}

	// Convert the user ID to a MongoDB ObjectID
	userPrimitiveID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to convert user ID to primitive.ObjectID")
		return
	}

	// Get the user name from the parameters or from the database
	userName := params.Name
	if userName == "" {
		for _, user := range users {
			if user.ID == userPrimitiveID {
				userName = user.Name
				break
			}
		}
	}

	// Get the user email from the parameters or from the database
	userEmail := params.Email
	if userEmail == "" {
		for _, user := range users {
			if user.ID == userPrimitiveID {
				userEmail = user.Email
				break
			}
		}
	}

	// Get the user password from the parameters or from the database
	var hashedPassword []byte
	userPassword := params.Password
	if userPassword == "" {
		for _, user := range users {
			if user.ID == userPrimitiveID {
				hashedPassword = []byte(user.Password)
				break
			}
		}
	} else {
		hashedPassword, err = bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Unable to hash password")
			return
		}
	}

	// Update the user in the database
	if err := client.UpdateUser(userID, userName, userEmail, string(hashedPassword)); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to update user")
		return
	}

	// Respond with a success message
	RespondWithJSON(w, http.StatusOK, "User updated successfully")
}

// DeleteUserHandler handles the HTTP request for deleting a user.
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the JWT token and the database client from the request.
	token, _, client := ExtractDBAndToken(r)

	// If the token is nil, return immediately.
	if token == nil {
		return
	}

	// Extract the issuer from the token claims.
	issuer := token.Claims.(jwt.MapClaims)["Issuer"].(string)

	// If the issuer is "go-chat-application-refresh", respond with an error
	// because a JWT access token is required, not a refresh token.
	if issuer == "go-chat-application-refresh" {
		RespondWithError(w, http.StatusUnauthorized,
			"Using JWT refresh token when JWT access token is required")
		return
	}

	// Extract the user ID from the token claims and remove the "ObjectID(" and ")" parts.
	userID := token.Claims.(jwt.MapClaims)["Subject"].(string)
	userID = strings.TrimPrefix(userID, "ObjectID(\"")
	userID = strings.TrimSuffix(userID, "\")")

	// If the user ID is empty, respond with an error.
	if userID == "" {
		RespondWithError(w, http.StatusUnauthorized, "Unable to get user ID from JWT token")
		return
	}

	// Try to delete the user with the given user ID. If an error occurs, respond with an error.
	if err := client.DeleteUser(userID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to delete user")
		return
	}

	// If everything went well, respond with a success message.
	RespondWithJSON(w, http.StatusOK, "User deleted successfully")
}

// LoginUserHandler handles user login requests
func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the database client from the request context
	ctx := r.Context()
	client, _ := ctx.Value(config.ApiCfg.DB).(*database.MongoDBClient)

	// Define the structure for the request parameters
	var params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body into the params structure
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Retrieve all users from the database
	users, err := client.GetAllUsers()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unable to get users")
		return
	}

	// Iterate over the users to find a match for the email
	for _, user := range users {
		if params.Email == user.Email {
			// Check if the provided password matches the user's password
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password)); err != nil {
				RespondWithError(w, http.StatusUnauthorized, "Invalid password")
			}

			// Generate UUIDs for the access and refresh tokens
			accessTokenID, err := uuid.NewUUID()
			if err != nil {
				RespondWithError(w, http.StatusInternalServerError, "Unable to generate access token")
				return
			}

			refreshTokenID, err := uuid.NewUUID()
			if err != nil {
				RespondWithError(w, http.StatusInternalServerError, "Unable to generate refresh token")
				return
			}

			// Define the claims for the access and refresh tokens
			accessClaims := jwt.MapClaims{
				"ID":        accessTokenID.String(),
				"Issuer":    "go-chat-application-access",
				"Subject":   user.ID.String(),
				"IssuedAt":  jwt.NewNumericDate(time.Now()),
				"ExpiresAt": jwt.NewNumericDate(time.Now().Add(AccessExpiration)),
				"Revoked":   false,
			}

			refreshClaims := jwt.MapClaims{
				"ID":        refreshTokenID.String(),
				"Issuer":    "go-chat-application-refresh",
				"Subject":   user.ID.String(),
				"IssuedAt":  jwt.NewNumericDate(time.Now()),
				"ExpiresAt": jwt.NewNumericDate(time.Now().Add(RefreshExpiration)),
				"Revoked":   false,
			}

			// Create the JWTs with the defined claims
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

			// Add the tokens to the token map
			tokenPackage.AddTokenToMap(token)
			tokenPackage.AddTokenToMap(refreshToken)

			// Generate a signed string from the access token using the JWT secret
			signedToken, err := token.SignedString([]byte(config.ApiCfg.JwtSecret))
			if err != nil {
				// If there's an error, respond with an internal server error message
				RespondWithError(w, http.StatusInternalServerError, "Unable to sign access token")
				return
			}

			// Generate a signed string from the refresh token using the JWT secret
			signedRefreshToken, err := refreshToken.SignedString([]byte(config.ApiCfg.JwtSecret))
			if err != nil {
				// If there's an error, respond with an internal server error message
				RespondWithError(w, http.StatusInternalServerError, "Unable to sign refresh token")
				return
			}

			// Create a map to hold the response data
			responseMap := map[string]interface{}{
				"id":            accessClaims["Subject"],
				"name":          user.Name,
				"email":         user.Email,
				"access_token":  signedToken,
				"refresh_token": signedRefreshToken,
			}

			// Respond with the created map as JSON
			RespondWithJSON(w, http.StatusOK, responseMap)
			return
		}
	}
	// If the function hasn't returned by this point, the email is invalid
	RespondWithError(w, http.StatusUnauthorized, "Invalid email")
}

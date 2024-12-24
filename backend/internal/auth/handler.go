package auth

import (
	"context"
	"fmt"

	"github.com/hossein-nas/analytics_aggregator/internal/config"
	pkg "github.com/hossein-nas/analytics_aggregator/pkg/responses"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	db            *mongo.Database
	prefix        string
	accessSecret  []byte
	refreshSecret []byte
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func NewHandler(db *mongo.Database, cfg config.JWTConfig, pathPrefix interface{}) *Handler {
	if pathPrefix == nil {
		pathPrefix = "/"
	}

	return &Handler{
		db:            db,
		prefix:        pathPrefix.(string),
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
	}
}

func (h *Handler) generateTokens(userID primitive.ObjectID, username string) (string, string, error) {
	// Access token - 15 minutes
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID.Hex(),
		"username": username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(h.accessSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh token - 7 days
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.Hex(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString(h.refreshSecret)
	if err != nil {
		return "", "", err
	}

	// Store refresh token in database
	rt := RefreshToken{
		UserID:    userID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Used:      false,
	}

	_, err = h.db.Collection("refresh_tokens").InsertOne(context.Background(), rt)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (h *Handler) setTokenCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     h.prefix + "/",
		MaxAge:   15 * 60, // 15 minutes
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     h.prefix + "/auth/refresh",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials User
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var user User
	err := h.db.Collection("users").FindOne(r.Context(), bson.M{
		"username": credentials.Username,
	}).Decode(&user)

	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.setTokenCookies(w, accessToken, refreshToken)

	pkg.RespondWithJSON(w, http.StatusOK, pkg.SuccessResponse{
		Message: "Login successful",
		Data: map[string]string{
			"user_id":  user.ID.Hex(),
			"username": user.Username,
		},
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.RespondWithError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	// Validate input
	if len(req.Username) < 3 {
		pkg.RespondWithError(w, http.StatusBadRequest, "Username must be at least 3 characters long")
		return
	}
	if len(req.Password) < 8 {
		pkg.RespondWithError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}

	// Check if username already exists
	var existingUser User
	err := h.db.Collection("users").FindOne(r.Context(), bson.M{
		"username": req.Username,
	}).Decode(&existingUser)

	if err == nil {
		pkg.RespondWithError(w, http.StatusConflict, "Username already exists")
		return
	}
	if err != mongo.ErrNoDocuments {
		pkg.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Create new user
	user := User{
		ID:       primitive.NewObjectID(),
		Username: req.Username,
		Password: string(hashedPassword),
	}

	// Insert user into database
	_, err = h.db.Collection("users").InsertOne(r.Context(), user)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set tokens in cookies
	h.setTokenCookies(w, accessToken, refreshToken)

	// Return success response
	response := RegisterResponse{
		ID:       user.ID.Hex(),
		Username: user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	pkg.RespondWithJSON(w, http.StatusOK, &pkg.SuccessResponse{
		Message: "User has registered successfully.",
		Data:    &response,
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refresh_token")
	fmt.Println(r.Cookies())
	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, "Refresh token not found")
		return
	}

	token, err := jwt.Parse(refreshCookie.Value, func(token *jwt.Token) (interface{}, error) {
		return h.refreshSecret, nil
	})

	if err != nil || !token.Valid {
		pkg.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token.")
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))

	// Find and invalidate the used refresh token
	var rt RefreshToken
	err = h.db.Collection("refresh_tokens").FindOne(r.Context(), bson.M{
		"user_id": userID,
		"token":   refreshCookie.Value,
		"used":    false,
	}).Decode(&rt)

	if err != nil {
		pkg.RespondWithError(w, http.StatusUnauthorized, "Invalid refresh token.")
		return
	}

	// Mark the current refresh token as used
	_, err = h.db.Collection("refresh_tokens").UpdateOne(
		r.Context(),
		bson.M{"_id": rt.ID},
		bson.M{"$set": bson.M{"used": true}},
	)

	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate new tokens
	var user User
	err = h.db.Collection("users").FindOne(r.Context(), bson.M{
		"_id": userID,
	}).Decode(&user)

	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		pkg.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.setTokenCookies(w, accessToken, refreshToken)
	w.WriteHeader(http.StatusOK)
	pkg.RespondWithJSON(w, http.StatusOK, &pkg.SuccessResponse{
		Message: "Refresh token and access token has updated.",
		Data:    nil,
	})
}

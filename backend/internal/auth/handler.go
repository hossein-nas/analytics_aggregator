package auth

import (
	"context"

	"github.com/hossein-nas/analytics_aggregator/internal/config"
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

func NewHandler(db *mongo.Database, cfg config.JWTConfig) *Handler {
	return &Handler{
		db:            db,
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
		Path:     "/",
		MaxAge:   15 * 60, // 15 minutes
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth/refresh",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials User
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var user User
	err := h.db.Collection("users").FindOne(r.Context(), bson.M{
		"username": credentials.Username,
	}).Decode(&user)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.setTokenCookies(w, accessToken, refreshToken)

	respondWithJSON(w, http.StatusOK, SuccessResponse{
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if len(req.Username) < 3 {
		http.Error(w, "Username must be at least 3 characters long", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	var existingUser User
	err := h.db.Collection("users").FindOne(r.Context(), bson.M{
		"username": req.Username,
	}).Decode(&existingUser)

	if err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}
	if err != mongo.ErrNoDocuments {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(refreshCookie.Value, func(token *jwt.Token) (interface{}, error) {
		return h.refreshSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
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
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Mark the current refresh token as used
	_, err = h.db.Collection("refresh_tokens").UpdateOne(
		r.Context(),
		bson.M{"_id": rt.ID},
		bson.M{"$set": bson.M{"used": true}},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate new tokens
	var user User
	err = h.db.Collection("users").FindOne(r.Context(), bson.M{
		"_id": userID,
	}).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accessToken, refreshToken, err := h.generateTokens(user.ID, user.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.setTokenCookies(w, accessToken, refreshToken)
	w.WriteHeader(http.StatusOK)
}

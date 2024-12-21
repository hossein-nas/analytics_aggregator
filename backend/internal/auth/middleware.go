package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserClaims struct {
	UserID   primitive.ObjectID
	Username string
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return h.accessSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert user_id to ObjectID
		userID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create user claims
		userClaims := UserClaims{
			UserID:   userID,
			Username: claims["username"].(string),
		}

		// Add user claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, userClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper function to get user claims from context
func GetUserFromContext(ctx context.Context) (UserClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(UserClaims)
	return claims, ok
}

// Optional: Bearer token support
func (h *Handler) extractBearerToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

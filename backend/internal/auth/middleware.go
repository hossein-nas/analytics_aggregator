package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	pkg "github.com/hossein-nas/analytics_aggregator/pkg/responses"
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
			pkg.RespondWithError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		// Parse token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return h.accessSecret, nil
		})

		if err != nil || !token.Valid {
			pkg.RespondWithError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			pkg.RespondWithError(w, http.StatusUnauthorized, "Unauthorized.")
			return
		}

		// Convert user_id to ObjectID
		userID, err := primitive.ObjectIDFromHex(claims["user_id"].(string))
		if err != nil {
			pkg.RespondWithError(w, http.StatusUnauthorized, "Unauthorized.")
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

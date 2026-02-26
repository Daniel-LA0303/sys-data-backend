package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("super_secret_key_change_me")
var JWTExpiration = time.Hour * 24 // 24 horas

type contextKey string

const UserIDKey contextKey = "userID"

func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

func WithJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := validateJWT(tokenString)
		if err != nil || !token.Valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok || userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		next(w, r.WithContext(ctx))
	}
}

func CreateJWT(userID string, orgId string, roleName string, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   userID,
		"orgId":    orgId,
		"roleName": roleName,
		"email":    email,
		"exp":      time.Now().Add(JWTExpiration).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		// Validar algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return JWTSecret, nil
	})
}

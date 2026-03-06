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

type CustomClaims struct {
	UserID   string `json:"userID"`
	Username string `json:"username"`
	OrgId    string `json:"orgId"`
	jwt.RegisteredClaims
}

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
			http.Error(w, "unauthorized: missing header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 1. Ahora ValidateJWT devuelve (*CustomClaims, error)
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 2. Ya no necesitas hacer cast a jwt.MapClaims
		// Accedemos directamente a la propiedad de la estructura
		userID := claims.UserID
		if userID == "" {
			http.Error(w, "unauthorized: empty user id", http.StatusUnauthorized)
			return
		}

		// 3. Inyectamos en el contexto
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		next(w, r.WithContext(ctx))
	}
}
func CreateJWT(userID string, orgId string, roleName string, email string, username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   userID,
		"orgId":    orgId,
		"roleName": roleName,
		"email":    email,
		"username": username,
		"exp":      time.Now().Add(JWTExpiration).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Hacemos el "type assertion" para obtener los claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token no válido o claims incorrectos")
}

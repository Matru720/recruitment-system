package auth

import (
	"context"
	"net/http"
	"recruitment-system/utils"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserTypeKey contextKey = "userType"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserTypeKey, claims.UserType)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userType, ok := r.Context().Value(UserTypeKey).(string)
		if !ok || userType != "Admin" {
			utils.RespondWithError(w, http.StatusForbidden, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ApplicantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userType, ok := r.Context().Value(UserTypeKey).(string)
		if !ok || userType != "Applicant" {
			utils.RespondWithError(w, http.StatusForbidden, "Applicant access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

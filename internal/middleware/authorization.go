package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	authz "github.com/wurt83ow/gophermart/internal/authorization"
	"github.com/wurt83ow/gophermart/internal/models"
	"go.uber.org/zap"
)

type Storage interface {
	InsertUser(k string, v models.DataUser) (models.DataUser, error)
}

// JWTAuthzMiddleware verifies a valid JWT exists in our
// cookie and if not, encourages the consumer to login again.
func JWTAuthzMiddleware(storage Storage, log Log) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// Grab jwt-token cookie
			jwtCookie, err := r.Cookie("jwt-token")

			var userID string
			if err == nil {
				userID, err = authz.DecodeJWTToUser(jwtCookie.Value)
				if err != nil {
					userID = ""
					log.Info("Error occurred creating a cookie", zap.Error(err))
				}

			} else {
				log.Info("Error occurred reading cookie", zap.Error(err))
			}

			if userID == "" {
				jwtCookie := r.Header.Get("Authorization")

				if jwtCookie != "" {
					userID, err = authz.DecodeJWTToUser(jwtCookie)

					if err != nil {
						userID = ""
						log.Info("Error occurred creating a cookie", zap.Error(err))
					}
				}
			}

			if userID == "" {
				userID = uuid.New().String()

				go func() {
					email := uuid.New().String()
					dataUser := models.DataUser{UUID: userID, Email: email, Name: "default"}
					_, err = storage.InsertUser(email, dataUser)
					if err != nil {
						log.Info("Error occurred user create", zap.Error(err))
					}
				}()

				freshToken := authz.CreateJWTTokenForUser(userID)
				http.SetCookie(w, authz.AuthCookie("jwt-token", freshToken))
				// http.SetCookie(w, authz.AuthCookie("Authorization", freshToken))
				w.Header().Set("Authorization", freshToken)
			}

			var keyUserID models.Key = "userID"
			ctx := r.Context()
			ctx = context.WithValue(ctx, keyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

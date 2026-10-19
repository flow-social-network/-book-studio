package auth

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type contextKey string

const userIDKey contextKey = "bookstudio.userID"

func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

func NewFirebaseVerifier(ctx context.Context, credentialsFile, projectID string) (*fbauth.Client, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	cfg := &firebase.Config{}
	if projectID != "" {
		cfg.ProjectID = projectID
	}
	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}
	return app.Auth(ctx)
}

func RequireUser(client *fbauth.Client, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			writeAuthError(w)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			writeAuthError(w)
			return
		}
		verified, err := client.VerifyIDToken(r.Context(), token)
		if err != nil || verified.UID == "" {
			writeAuthError(w)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, verified.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeAuthError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":{"code":"UNAUTHENTICATED","message":"A valid Firebase ID token is required."}}`))
}

package middleware

import (
	"context"

	"net/http"
	"strings"

	"github.com/bluekeyes/hatpear"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

type userCtxKey struct{}

type TokenResolver interface {
	// ResolveToken resolves a token to a username and set of scopes. If the
	// token is not valid, it returns an error.
	ResolveToken(ctx context.Context, token string) (user string, scopes []string, err error)
}

// APIAuth returns middleware that rejects requests if they do not include a
// valid GitHub token with 'repo' scope. It stores the name of the
// authenticated user (but not their token) in the request context.
func APIAuth(resolver TokenResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return hatpear.TryFunc(func(w http.ResponseWriter, r *http.Request) error {
			ctx := r.Context()

			token := getBearerToken(r)
			if token == "" {
				return errors.New("missing token")
			}

			user, scopes, err := resolver.ResolveToken(ctx, token)
			if err != nil {
				return errors.Wrap(err, "failed to resolve token")
			}

			if hasScope(scopes, "repo") {
				ctx := context.WithValue(ctx, userCtxKey{}, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return nil
			}

			return errors.New("token must have 'repo' scope")
		})
	}
}

type GitHubTokenResolver struct {
	ClientCreator githubapp.ClientCreator
}

func (r *GitHubTokenResolver) ResolveToken(ctx context.Context, token string) (string, []string, error) {
	client, err := r.ClientCreator.NewTokenClient(token)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to create token client")
	}

	user, res, err := client.Users.Get(ctx, "")
	if err != nil {
		if res != nil {
			switch res.StatusCode {
			case http.StatusUnauthorized, http.StatusForbidden:
				return "", nil, errors.Wrap(err, "invalid token")
			}
		}
		return "", nil, errors.Wrap(err, "failed to get authenticating user")
	}

	var scopes []string
	for _, scope := range strings.Split(res.Header.Get("X-OAuth-Scopes"), ",") {
		scopes = append(scopes, strings.TrimSpace(scope))
	}

	return user.GetLogin(), scopes, nil
}

func hasScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}

	return false
}

func getBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if token, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return token
	}

	return ""
}

package ssh

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	gossh "golang.org/x/crypto/ssh"
)

type PublicKeyContextKey string

const PublicKeyKey PublicKeyContextKey = "public_key"

func PublicKeyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	pubKeyStr := fmt.Sprintf("%s %s", key.Type(), gossh.FingerprintSHA256(key))

	ctx.SetValue(string(PublicKeyKey), pubKeyStr)

	log.Info("Public key auth", "user", ctx.User(), "public_key", pubKeyStr)

	return true
}

func WithPublicKeyMiddleware() wish.Middleware {
	return func(handler ssh.Handler) ssh.Handler {
		return func(session ssh.Session) {
			handler(session)
		}
	}
}

func GetPublicKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	v := ctx.Value(string(PublicKeyKey))
	if v == nil {
		return ""
	}

	pubKey, ok := v.(string)
	if !ok {
		return ""
	}

	return pubKey
}

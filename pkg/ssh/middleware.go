package ssh

import (
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func AuthenticationMiddleware(sh ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		pk := s.PublicKey()
		log.Info("Public Key", pk)
	}
}

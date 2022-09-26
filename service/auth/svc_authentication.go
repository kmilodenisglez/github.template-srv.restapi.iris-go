package auth

import (
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/repo/db"
)

type SvcAuthentication struct {
	AuthProviders map[string]Provider // similar to slices, maps are reference types.
}

// NewSvcAuthentication creates the authentication service. It provides the methods to make the
// authentication intent with the register providers.
//
// - providers [Array] ~ Maps of providers string token / identifiers
//
// - conf [*SvcConfig] ~ App conf instance pointer
func NewSvcAuthentication(providers map[string]bool, repoUser *db.RepoDrones) *SvcAuthentication {
	k := &SvcAuthentication{AuthProviders: make(map[string]Provider)}

	for v := range providers {
		k.AuthProviders[v] = &ProviderDrone{
			repo: repoUser,
		}
	}

	return k
}

package auth

import (
	"net/url"
	"path"

	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
)

type EndpointsProvider struct {
	PrefixProvider urlprefix.Provider
}

func (p *EndpointsProvider) urlOf(relPath string) *url.URL {
	u := p.PrefixProvider.Value()
	u.Path = path.Join(u.Path, relPath)
	return u
}

func (p *EndpointsProvider) AuthorizeEndpointURI() *url.URL    { return p.urlOf("oauth2/authorize") }
func (p *EndpointsProvider) TokenEndpointURI() *url.URL        { return p.urlOf("oauth2/token") }
func (p *EndpointsProvider) JWKSEndpointURI() *url.URL         { return p.urlOf("oauth2/jwks") }
func (p *EndpointsProvider) AuthenticateEndpointURI() *url.URL { return p.urlOf(".") }

var endpointsProviderSet = wire.NewSet(
	wire.Struct(new(EndpointsProvider), "*"),
	wire.Bind(new(oauth.AuthorizeEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oauth.TokenEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oidc.JWKSEndpointProvider), new(*EndpointsProvider)),
	wire.Bind(new(oauth.AuthenticateEndpointProvider), new(*EndpointsProvider)),
)

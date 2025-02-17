package adfs

import (
	"context"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	liboauthrelyingparty "github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	oauthrelyingparty.RegisterProvider(Type, ADFS{})
}

const Type = liboauthrelyingparty.TypeADFS

type ProviderConfig oauthrelyingparty.ProviderConfig

func (c ProviderConfig) DiscoveryDocumentEndpoint() string {
	discovery_document_endpoint, _ := c["discovery_document_endpoint"].(string)
	return discovery_document_endpoint
}

var _ oauthrelyingparty.Provider = ADFS{}
var _ liboauthrelyingparty.BuiltinProvider = ADFS{}

var Schema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"alias": { "type": "string" },
		"type": { "type": "string" },
		"modify_disabled": { "type": "boolean" },
		"client_id": { "type": "string", "minLength": 1 },
		"claims": {
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"email": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"assume_verified": { "type": "boolean" },
						"required": { "type": "boolean" }
					}
				}
			}
		},
		"discovery_document_endpoint": { "type": "string", "format": "uri" }
	},
	"required": ["alias", "type", "client_id", "discovery_document_endpoint"]
}
`)

type ADFS struct{}

func (ADFS) ValidateProviderConfig(ctx *validation.Context, cfg oauthrelyingparty.ProviderConfig) {
	ctx.AddError(Schema.Validator().ValidateValue(cfg))
}

func (ADFS) SetDefaults(cfg oauthrelyingparty.ProviderConfig) {
	cfg.SetDefaultsModifyDisabledFalse()
	cfg.SetDefaultsEmailClaimConfig(oauthrelyingpartyutil.Email_AssumeVerified_Required())
}

func (ADFS) ProviderID(cfg oauthrelyingparty.ProviderConfig) oauthrelyingparty.ProviderID {
	// In the original implementation, provider ID is just type.
	return oauthrelyingparty.NewProviderID(cfg.Type(), nil)
}

func (ADFS) scope() []string {
	// The supported scopes are observed from a AD FS server.
	return []string{"openid", "profile", "email"}
}

func (ADFS) getOpenIDConfiguration(deps oauthrelyingparty.Dependencies) (*oauthrelyingpartyutil.OIDCDiscoveryDocument, error) {
	endpoint := ProviderConfig(deps.ProviderConfig).DiscoveryDocumentEndpoint()
	return oauthrelyingpartyutil.FetchOIDCDiscoveryDocument(deps.HTTPClient, endpoint)
}

func (p ADFS) GetAuthorizationURL(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetAuthorizationURLOptions) (string, error) {
	c, err := p.getOpenIDConfiguration(deps)
	if err != nil {
		return "", err
	}
	return c.MakeOAuthURL(oauthrelyingpartyutil.AuthorizationURLParams{
		ClientID:     deps.ProviderConfig.ClientID(),
		RedirectURI:  param.RedirectURI,
		Scope:        p.scope(),
		ResponseType: oauthrelyingparty.ResponseTypeCode,
		ResponseMode: param.ResponseMode,
		State:        param.State,
		Prompt:       p.getPrompt(param.Prompt),
		Nonce:        param.Nonce,
	}), nil
}

func (p ADFS) GetUserProfile(deps oauthrelyingparty.Dependencies, param oauthrelyingparty.GetUserProfileOptions) (authInfo oauthrelyingparty.UserProfile, err error) {
	c, err := p.getOpenIDConfiguration(deps)
	if err != nil {
		return
	}

	// OPTIMIZE(sso): Cache JWKs
	keySet, err := c.FetchJWKs(deps.HTTPClient)
	if err != nil {
		return
	}

	var tokenResp oauthrelyingpartyutil.AccessTokenResp
	jwtToken, err := c.ExchangeCode(
		deps.HTTPClient,
		deps.Clock,
		param.Code,
		keySet,
		deps.ProviderConfig.ClientID(),
		deps.ClientSecret,
		param.RedirectURI,
		param.Nonce,
		&tokenResp,
	)
	if err != nil {
		return
	}

	claims, err := jwtToken.AsMap(context.TODO())
	if err != nil {
		return
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("sub not found in ID token")
		return
	}

	// The upn claim is documented here.
	// https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/configuring-alternate-login-id
	upn, ok := claims["upn"].(string)
	if !ok {
		err = oauthrelyingpartyutil.OAuthProtocolError.New("upn not found in ID token")
		return
	}

	extracted, err := stdattrs.Extract(claims, stdattrs.ExtractOptions{})
	if err != nil {
		return
	}

	// Transform upn into preferred_username
	if _, ok := extracted[stdattrs.PreferredUsername]; !ok {
		extracted[stdattrs.PreferredUsername] = upn
	}
	// Transform upn into email
	if _, ok := extracted[stdattrs.Email]; !ok {
		if emailErr := (validation.FormatEmail{}).CheckFormat(upn); emailErr == nil {
			// upn looks like an email address.
			extracted[stdattrs.Email] = upn
		}
	}

	emailRequired := deps.ProviderConfig.EmailClaimConfig().Required()
	extracted, err = stdattrs.Extract(extracted, stdattrs.ExtractOptions{
		EmailRequired: emailRequired,
	})
	if err != nil {
		return
	}
	authInfo.StandardAttributes = extracted

	authInfo.ProviderRawProfile = claims
	authInfo.ProviderUserID = sub

	return
}

func (ADFS) getPrompt(prompt []string) []string {
	// ADFS only supports prompt=login
	// https://docs.microsoft.com/en-us/windows-server/identity/ad-fs/operations/ad-fs-prompt-login
	for _, p := range prompt {
		if p == "login" {
			return []string{"login"}
		}
	}
	return []string{}
}

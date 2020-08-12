package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	authredis "github.com/authgear/authgear-server/pkg/auth/dependency/auth/redis"
	authenticatoroob "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/oob"
	authenticatorpassword "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/password"
	authenticatorservice "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/service"
	authenticatortotp "github.com/authgear/authgear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/auth/dependency/challenge"
	identityanonymous "github.com/authgear/authgear-server/pkg/auth/dependency/identity/anonymous"
	identityloginid "github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/auth/dependency/identity/oauth"
	identityservice "github.com/authgear/authgear-server/pkg/auth/dependency/identity/service"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	oauthhandler "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/handler"
	oauthpq "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/pq"
	oauthredis "github.com/authgear/authgear-server/pkg/auth/dependency/oauth/redis"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oidc"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	sessionredis "github.com/authgear/authgear-server/pkg/auth/dependency/session/redis"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/auth/dependency/user"
	"github.com/authgear/authgear-server/pkg/auth/dependency/webapp"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/feature/forgotpassword"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/feature/welcomemessage"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	taskqueue "github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
	"github.com/authgear/authgear-server/pkg/mfa"
	"github.com/authgear/authgear-server/pkg/otp"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var commonDeps = wire.NewSet(
	configDeps,
	utilsDeps,

	clock.DependencySet,
	db.DependencySet,

	wire.NewSet(
		challenge.DependencySet,
		wire.Bind(new(newinteraction.ChallengeProvider), new(*challenge.Provider)),
	),

	wire.NewSet(
		taskqueue.DependencySet,
		wire.Bind(new(task.Queue), new(*taskqueue.Queue)),
	),

	wire.NewSet(
		hook.DependencySet,
		wire.Bind(new(newinteraction.HookProvider), new(*hook.Provider)),
		wire.Bind(new(user.HookProvider), new(*hook.Provider)),
		wire.Bind(new(auth.HookProvider), new(*hook.Provider)),
	),

	wire.NewSet(
		sessionredis.DependencySet,
		wire.Bind(new(session.Store), new(*sessionredis.Store)),

		session.DependencySet,
		wire.Bind(new(auth.IDPSessionResolver), new(*session.Resolver)),
		wire.Bind(new(auth.IDPSessionManager), new(*session.Manager)),
		wire.Bind(new(oauth.ResolverSessionProvider), new(*session.Provider)),
		wire.Bind(new(oauthhandler.SessionProvider), new(*session.Provider)),
		wire.Bind(new(newinteraction.SessionProvider), new(*session.Provider)),
	),

	wire.NewSet(
		authredis.DependencySet,
		wire.Bind(new(auth.AccessEventStore), new(*authredis.EventStore)),

		auth.DependencySet,
		wire.Bind(new(session.AccessEventProvider), new(*auth.AccessEventProvider)),
	),

	wire.NewSet(
		authenticatorpassword.DependencySet,
		authenticatoroob.DependencySet,
		wire.Bind(new(newinteraction.OOBAuthenticatorProvider), new(*authenticatoroob.Provider)),
		authenticatortotp.DependencySet,

		authenticatorservice.DependencySet,
		wire.Bind(new(authenticatorservice.PasswordAuthenticatorProvider), new(*authenticatorpassword.Provider)),
		wire.Bind(new(authenticatorservice.OOBOTPAuthenticatorProvider), new(*authenticatoroob.Provider)),
		wire.Bind(new(authenticatorservice.TOTPAuthenticatorProvider), new(*authenticatortotp.Provider)),

		wire.Bind(new(forgotpassword.AuthenticatorService), new(*authenticatorservice.Service)),
		wire.Bind(new(newinteraction.AuthenticatorService), new(*authenticatorservice.Service)),
		wire.Bind(new(verification.AuthenticatorService), new(*authenticatorservice.Service)),
	),

	wire.NewSet(
		mfa.DependencySet,

		wire.Bind(new(newinteraction.MFAService), new(*mfa.Service)),
	),

	wire.NewSet(
		identityloginid.DependencySet,
		wire.Bind(new(sso.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		wire.Bind(new(newinteraction.LoginIDNormalizerFactory), new(*identityloginid.NormalizerFactory)),
		identityoauth.DependencySet,
		identityanonymous.DependencySet,
		wire.Bind(new(webapp.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(newinteraction.AnonymousIdentityProvider), new(*identityanonymous.Provider)),

		identityservice.DependencySet,
		wire.Bind(new(identityservice.LoginIDIdentityProvider), new(*identityloginid.Provider)),
		wire.Bind(new(identityservice.OAuthIdentityProvider), new(*identityoauth.Provider)),
		wire.Bind(new(identityservice.AnonymousIdentityProvider), new(*identityanonymous.Provider)),
		wire.Bind(new(user.IdentityService), new(*identityservice.Service)),
		wire.Bind(new(newinteraction.IdentityService), new(*identityservice.Service)),
		wire.Bind(new(forgotpassword.IdentityService), new(*identityservice.Service)),
	),

	wire.NewSet(
		user.DependencySet,
		wire.Bind(new(auth.UserProvider), new(*user.Queries)),
		wire.Bind(new(newinteraction.UserService), new(*user.Provider)),
		wire.Bind(new(oidc.UserProvider), new(*user.Queries)),
		wire.Bind(new(hook.UserProvider), new(*user.RawProvider)),
	),

	wire.NewSet(
		sso.DependencySet,
		wire.Bind(new(newinteraction.OAuthProviderFactory), new(*sso.OAuthProviderFactory)),
	),

	wire.NewSet(
		welcomemessage.DependencySet,
		wire.Bind(new(user.WelcomeMessageProvider), new(*welcomemessage.Provider)),
	),

	wire.NewSet(
		forgotpassword.DependencySet,
		wire.Bind(new(newinteraction.ForgotPasswordService), new(*forgotpassword.Provider)),
		wire.Bind(new(newinteraction.ResetPasswordService), new(*forgotpassword.Provider)),
	),

	wire.NewSet(
		oauthpq.DependencySet,
		wire.Bind(new(oauth.AuthorizationStore), new(*oauthpq.AuthorizationStore)),

		oauthredis.DependencySet,
		wire.Bind(new(oauth.AccessGrantStore), new(*oauthredis.GrantStore)),
		wire.Bind(new(oauth.CodeGrantStore), new(*oauthredis.GrantStore)),
		wire.Bind(new(oauth.OfflineGrantStore), new(*oauthredis.GrantStore)),

		oauth.DependencySet,
		wire.Bind(new(auth.AccessTokenSessionResolver), new(*oauth.Resolver)),
		wire.Bind(new(auth.AccessTokenSessionManager), new(*oauth.SessionManager)),
		wire.Bind(new(oauthhandler.OAuthURLProvider), new(*oauth.URLProvider)),
		wire.Value(oauthhandler.TokenGenerator(oauth.GenerateToken)),
	),

	wire.NewSet(
		oidc.DependencySet,
		wire.Value(oauthhandler.ScopesValidator(oidc.ValidateScopes)),
		wire.Bind(new(oauthhandler.IDTokenIssuer), new(*oidc.IDTokenIssuer)),
	),

	wire.NewSet(
		newinteraction.DependencySet,
		wire.Bind(new(webapp.GraphService), new(*newinteraction.Service)),
		wire.Bind(new(oauthhandler.GraphService), new(*newinteraction.Service)),
	),

	wire.NewSet(
		endpoints.DependencySet,
		wire.Bind(new(oauth.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(webapp.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(handlerwebapp.SetupTOTPEndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(authenticatoroob.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(oidc.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(sso.EndpointsProvider), new(*endpoints.Provider)),
		wire.Bind(new(otp.EndpointsProvider), new(*endpoints.Provider)),
	),

	wire.NewSet(
		verification.DependencySet,
		wire.Bind(new(user.VerificationService), new(*verification.Service)),
		wire.Bind(new(newinteraction.VerificationService), new(*verification.Service)),
	),

	wire.NewSet(
		otp.DependencySet,
		wire.Bind(new(authenticatoroob.OTPMessageSender), new(*otp.MessageSender)),
		wire.Bind(new(verification.OTPMessageSender), new(*otp.MessageSender)),
	),
)

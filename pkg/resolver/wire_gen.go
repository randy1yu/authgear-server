// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package resolver

import (
	"context"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/oob"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	service2 "github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/totp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/biometric"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/event"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/feature/welcomemessage"
	"github.com/authgear/authgear-server/pkg/lib/healthz"
	"github.com/authgear/authgear-server/pkg/lib/hook"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
	oauth2 "github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oidc"
	"github.com/authgear/authgear-server/pkg/lib/oauth/pq"
	"github.com/authgear/authgear-server/pkg/lib/oauth/redis"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/resolver/handler"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/template"
	"net/http"
)

// Injectors from wire.go:

func newHealthzHandler(p *deps.RootProvider, w http.ResponseWriter, r *http.Request, ctx context.Context) http.Handler {
	pool := p.DatabasePool
	environmentConfig := p.EnvironmentConfig
	databaseEnvironmentConfig := &environmentConfig.Database
	factory := p.LoggerFactory
	handle := globaldb.NewHandle(ctx, pool, databaseEnvironmentConfig, factory)
	sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
	handlerLogger := healthz.NewHandlerLogger(factory)
	handler := &healthz.Handler{
		Context:        ctx,
		GlobalDatabase: handle,
		GlobalExecutor: sqlExecutor,
		Logger:         handlerLogger,
	}
	return handler
}

func newSentryMiddleware(p *deps.RootProvider) httproute.Middleware {
	hub := p.SentryHub
	environmentConfig := p.EnvironmentConfig
	trustProxy := environmentConfig.TrustProxy
	sentryMiddleware := &middleware.SentryMiddleware{
		SentryHub:  hub,
		TrustProxy: trustProxy,
	}
	return sentryMiddleware
}

func newPanicEndMiddleware(p *deps.RootProvider) httproute.Middleware {
	panicEndMiddleware := &middleware.PanicEndMiddleware{}
	return panicEndMiddleware
}

func newPanicWriteEmptyResponseMiddleware(p *deps.RootProvider) httproute.Middleware {
	panicWriteEmptyResponseMiddleware := &middleware.PanicWriteEmptyResponseMiddleware{}
	return panicWriteEmptyResponseMiddleware
}

func newBodyLimitMiddleware(p *deps.RootProvider) httproute.Middleware {
	bodyLimitMiddleware := &middleware.BodyLimitMiddleware{}
	return bodyLimitMiddleware
}

func newPanicLogMiddleware(p *deps.RequestProvider) httproute.Middleware {
	appProvider := p.AppProvider
	factory := appProvider.LoggerFactory
	panicLogMiddlewareLogger := middleware.NewPanicLogMiddlewareLogger(factory)
	panicLogMiddleware := &middleware.PanicLogMiddleware{
		Logger: panicLogMiddlewareLogger,
	}
	return panicLogMiddleware
}

func newSessionMiddleware(p *deps.RequestProvider) httproute.Middleware {
	appProvider := p.AppProvider
	config := appProvider.Config
	appConfig := config.AppConfig
	sessionConfig := appConfig.Session
	cookieDef := session.NewSessionCookieDef(sessionConfig)
	request := p.Request
	rootProvider := appProvider.RootProvider
	environmentConfig := rootProvider.EnvironmentConfig
	trustProxy := environmentConfig.TrustProxy
	httpConfig := appConfig.HTTP
	cookieManager := deps.NewCookieManager(request, trustProxy, httpConfig)
	contextContext := deps.ProvideRequestContext(request)
	appID := appConfig.ID
	handle := appProvider.Redis
	clock := _wireSystemClockValue
	factory := appProvider.LoggerFactory
	storeRedisLogger := idpsession.NewStoreRedisLogger(factory)
	storeRedis := &idpsession.StoreRedis{
		Redis:  handle,
		AppID:  appID,
		Clock:  clock,
		Logger: storeRedisLogger,
	}
	eventStoreRedis := &access.EventStoreRedis{
		Redis: handle,
		AppID: appID,
	}
	eventProvider := &access.EventProvider{
		Store: eventStoreRedis,
	}
	rand := _wireRandValue
	provider := &idpsession.Provider{
		Context:      contextContext,
		Request:      request,
		AppID:        appID,
		Redis:        handle,
		Store:        storeRedis,
		AccessEvents: eventProvider,
		TrustProxy:   trustProxy,
		Config:       sessionConfig,
		Clock:        clock,
		Random:       rand,
	}
	resolver := &idpsession.Resolver{
		Cookies:    cookieManager,
		CookieDef:  cookieDef,
		Provider:   provider,
		TrustProxy: trustProxy,
		Clock:      clock,
	}
	oAuthConfig := appConfig.OAuth
	secretConfig := config.SecretConfig
	databaseCredentials := deps.ProvideDatabaseCredentials(secretConfig)
	sqlBuilder := appdb.NewSQLBuilder(databaseCredentials, appID)
	appdbHandle := appProvider.AppDatabase
	sqlExecutor := appdb.NewSQLExecutor(contextContext, appdbHandle)
	authorizationStore := &pq.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	logger := redis.NewLogger(factory)
	store := &redis.Store{
		Context:     contextContext,
		Redis:       handle,
		AppID:       appID,
		Logger:      logger,
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
		Clock:       clock,
	}
	oAuthKeyMaterials := deps.ProvideOAuthKeyMaterials(secretConfig)
	endpointsProvider := &EndpointsProvider{
		HTTP: httpConfig,
	}
	userStore := &user.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	authenticationConfig := appConfig.Authentication
	identityConfig := appConfig.Identity
	featureConfig := config.FeatureConfig
	identityFeatureConfig := featureConfig.Identity
	serviceStore := &service.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	loginidStore := &loginid.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	loginIDConfig := identityConfig.LoginID
	manager := appProvider.Resources
	typeCheckerFactory := &loginid.TypeCheckerFactory{
		Config:    loginIDConfig,
		Resources: manager,
	}
	checker := &loginid.Checker{
		Config:             loginIDConfig,
		TypeCheckerFactory: typeCheckerFactory,
	}
	normalizerFactory := &loginid.NormalizerFactory{
		Config: loginIDConfig,
	}
	loginidProvider := &loginid.Provider{
		Store:             loginidStore,
		Config:            loginIDConfig,
		Checker:           checker,
		NormalizerFactory: normalizerFactory,
		Clock:             clock,
	}
	oauthStore := &oauth.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	oauthProvider := &oauth.Provider{
		Store: oauthStore,
		Clock: clock,
	}
	anonymousStore := &anonymous.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	anonymousProvider := &anonymous.Provider{
		Store: anonymousStore,
		Clock: clock,
	}
	biometricStore := &biometric.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	biometricProvider := &biometric.Provider{
		Store: biometricStore,
		Clock: clock,
	}
	serviceService := &service.Service{
		Authentication:        authenticationConfig,
		Identity:              identityConfig,
		IdentityFeatureConfig: identityFeatureConfig,
		Store:                 serviceStore,
		LoginID:               loginidProvider,
		OAuth:                 oauthProvider,
		Anonymous:             anonymousProvider,
		Biometric:             biometricProvider,
	}
	store2 := &service2.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	passwordStore := &password.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	authenticatorConfig := appConfig.Authenticator
	authenticatorPasswordConfig := authenticatorConfig.Password
	passwordLogger := password.NewLogger(factory)
	historyStore := &password.HistoryStore{
		Clock:       clock,
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	passwordChecker := password.ProvideChecker(authenticatorPasswordConfig, historyStore)
	housekeeperLogger := password.NewHousekeeperLogger(factory)
	housekeeper := &password.Housekeeper{
		Store:  historyStore,
		Logger: housekeeperLogger,
		Config: authenticatorPasswordConfig,
	}
	passwordProvider := &password.Provider{
		Store:           passwordStore,
		Config:          authenticatorPasswordConfig,
		Clock:           clock,
		Logger:          passwordLogger,
		PasswordHistory: historyStore,
		PasswordChecker: passwordChecker,
		Housekeeper:     housekeeper,
	}
	totpStore := &totp.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	authenticatorTOTPConfig := authenticatorConfig.TOTP
	totpProvider := &totp.Provider{
		Store:  totpStore,
		Config: authenticatorTOTPConfig,
		Clock:  clock,
	}
	authenticatorOOBConfig := authenticatorConfig.OOB
	oobStore := &oob.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	oobStoreRedis := &oob.StoreRedis{
		Redis: handle,
		AppID: appID,
		Clock: clock,
	}
	oobLogger := oob.NewLogger(factory)
	oobProvider := &oob.Provider{
		Config:    authenticatorOOBConfig,
		Store:     oobStore,
		CodeStore: oobStoreRedis,
		Clock:     clock,
		Logger:    oobLogger,
	}
	ratelimitLogger := ratelimit.NewLogger(factory)
	storageRedis := &ratelimit.StorageRedis{
		AppID: appID,
		Redis: handle,
	}
	limiter := &ratelimit.Limiter{
		Logger:  ratelimitLogger,
		Storage: storageRedis,
		Clock:   clock,
	}
	service3 := &service2.Service{
		Store:       store2,
		Password:    passwordProvider,
		TOTP:        totpProvider,
		OOBOTP:      oobProvider,
		RateLimiter: limiter,
	}
	verificationLogger := verification.NewLogger(factory)
	verificationConfig := appConfig.Verification
	verificationStoreRedis := &verification.StoreRedis{
		Redis: handle,
		AppID: appID,
		Clock: clock,
	}
	storePQ := &verification.StorePQ{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	verificationService := &verification.Service{
		Request:     request,
		Logger:      verificationLogger,
		Config:      verificationConfig,
		TrustProxy:  trustProxy,
		Clock:       clock,
		CodeStore:   verificationStoreRedis,
		ClaimStore:  storePQ,
		RateLimiter: limiter,
	}
	storeDeviceTokenRedis := &mfa.StoreDeviceTokenRedis{
		Redis: handle,
		AppID: appID,
		Clock: clock,
	}
	storeRecoveryCodePQ := &mfa.StoreRecoveryCodePQ{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	mfaService := &mfa.Service{
		DeviceTokens:  storeDeviceTokenRedis,
		RecoveryCodes: storeRecoveryCodePQ,
		Clock:         clock,
		Config:        authenticationConfig,
		RateLimiter:   limiter,
	}
	defaultLanguageTag := deps.ProvideDefaultLanguageTag(config)
	supportedLanguageTags := deps.ProvideSupportedLanguageTags(config)
	templateResolver := &template.Resolver{
		Resources:             manager,
		DefaultLanguageTag:    defaultLanguageTag,
		SupportedLanguageTags: supportedLanguageTags,
	}
	engine := &template.Engine{
		Resolver: templateResolver,
	}
	localizationConfig := appConfig.Localization
	staticAssetURLPrefix := environmentConfig.StaticAssetURLPrefix
	staticAssetResolver := &web.StaticAssetResolver{
		Context:            contextContext,
		Config:             httpConfig,
		Localization:       localizationConfig,
		StaticAssetsPrefix: staticAssetURLPrefix,
		Resources:          manager,
	}
	translationService := &translation.Service{
		Context:        contextContext,
		TemplateEngine: engine,
		StaticAssets:   staticAssetResolver,
	}
	welcomeMessageConfig := appConfig.WelcomeMessage
	queue := appProvider.TaskQueue
	eventLogger := event.NewLogger(factory)
	storeImpl := &event.StoreImpl{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	hookLogger := hook.NewLogger(factory)
	hookConfig := appConfig.Hook
	webhookKeyMaterials := deps.ProvideWebhookKeyMaterials(secretConfig)
	syncHTTPClient := hook.NewSyncHTTPClient(hookConfig)
	asyncHTTPClient := hook.NewAsyncHTTPClient()
	deliverer := &hook.Deliverer{
		Config:    hookConfig,
		Secret:    webhookKeyMaterials,
		Clock:     clock,
		SyncHTTP:  syncHTTPClient,
		AsyncHTTP: asyncHTTPClient,
	}
	sink := &hook.Sink{
		Logger:    hookLogger,
		Deliverer: deliverer,
	}
	auditLogger := audit.NewLogger(factory)
	writeHandle := appProvider.AuditWriteDatabase
	auditDatabaseCredentials := deps.ProvideAuditDatabaseCredentials(secretConfig)
	auditdbSQLBuilder := auditdb.NewSQLBuilder(auditDatabaseCredentials, appID)
	writeSQLExecutor := auditdb.NewWriteSQLExecutor(contextContext, writeHandle)
	writeStore := &audit.WriteStore{
		SQLBuilder:  auditdbSQLBuilder,
		SQLExecutor: writeSQLExecutor,
	}
	auditSink := &audit.Sink{
		Logger:   auditLogger,
		Database: writeHandle,
		Store:    writeStore,
	}
	eventService := event.NewService(contextContext, request, trustProxy, eventLogger, appdbHandle, clock, localizationConfig, storeImpl, sink, auditSink)
	welcomemessageProvider := &welcomemessage.Provider{
		Translation:          translationService,
		RateLimiter:          limiter,
		WelcomeMessageConfig: welcomeMessageConfig,
		TaskQueue:            queue,
		Events:               eventService,
	}
	rawCommands := &user.RawCommands{
		Store:                  userStore,
		Clock:                  clock,
		WelcomeMessageProvider: welcomemessageProvider,
	}
	idpsessionManager := &idpsession.Manager{
		Store:     storeRedis,
		Clock:     clock,
		Config:    sessionConfig,
		Cookies:   cookieManager,
		CookieDef: cookieDef,
	}
	sessionManager := &oauth2.SessionManager{
		Store:  store,
		Clock:  clock,
		Config: oAuthConfig,
	}
	coordinator := &facade.Coordinator{
		Identities:      serviceService,
		Authenticators:  service3,
		Verification:    verificationService,
		MFA:             mfaService,
		Users:           rawCommands,
		PasswordHistory: historyStore,
		OAuth:           authorizationStore,
		IDPSessions:     idpsessionManager,
		OAuthSessions:   sessionManager,
		IdentityConfig:  identityConfig,
	}
	identityFacade := facade.IdentityFacade{
		Coordinator: coordinator,
	}
	authenticatorFacade := facade.AuthenticatorFacade{
		Coordinator: coordinator,
	}
	queries := &user.Queries{
		Store:          userStore,
		Identities:     identityFacade,
		Authenticators: authenticatorFacade,
		Verification:   verificationService,
	}
	idTokenIssuer := &oidc.IDTokenIssuer{
		Secrets: oAuthKeyMaterials,
		BaseURL: endpointsProvider,
		Users:   queries,
		Clock:   clock,
	}
	accessTokenEncoding := &oauth2.AccessTokenEncoding{
		Secrets:    oAuthKeyMaterials,
		Clock:      clock,
		UserClaims: idTokenIssuer,
		BaseURL:    endpointsProvider,
	}
	oauthResolver := &oauth2.Resolver{
		OAuthConfig:        oAuthConfig,
		TrustProxy:         trustProxy,
		Authorizations:     authorizationStore,
		AccessGrants:       store,
		OfflineGrants:      store,
		AppSessions:        store,
		AccessTokenDecoder: accessTokenEncoding,
		Sessions:           provider,
		Cookies:            cookieManager,
		Clock:              clock,
	}
	middlewareLogger := session.NewMiddlewareLogger(factory)
	sessionMiddleware := &session.Middleware{
		SessionCookie:              cookieDef,
		Cookies:                    cookieManager,
		IDPSessionResolver:         resolver,
		AccessTokenSessionResolver: oauthResolver,
		AccessEvents:               eventProvider,
		Users:                      queries,
		Database:                   appdbHandle,
		Logger:                     middlewareLogger,
	}
	return sessionMiddleware
}

var (
	_wireSystemClockValue = clock.NewSystemClock()
	_wireRandValue        = idpsession.Rand(rand.SecureRand)
)

func newSessionResolveHandler(p *deps.RequestProvider) http.Handler {
	appProvider := p.AppProvider
	handle := appProvider.AppDatabase
	config := appProvider.Config
	appConfig := config.AppConfig
	authenticationConfig := appConfig.Authentication
	identityConfig := appConfig.Identity
	featureConfig := config.FeatureConfig
	identityFeatureConfig := featureConfig.Identity
	secretConfig := config.SecretConfig
	databaseCredentials := deps.ProvideDatabaseCredentials(secretConfig)
	appID := appConfig.ID
	sqlBuilder := appdb.NewSQLBuilder(databaseCredentials, appID)
	request := p.Request
	contextContext := deps.ProvideRequestContext(request)
	sqlExecutor := appdb.NewSQLExecutor(contextContext, handle)
	store := &service.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	loginidStore := &loginid.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	loginIDConfig := identityConfig.LoginID
	manager := appProvider.Resources
	typeCheckerFactory := &loginid.TypeCheckerFactory{
		Config:    loginIDConfig,
		Resources: manager,
	}
	checker := &loginid.Checker{
		Config:             loginIDConfig,
		TypeCheckerFactory: typeCheckerFactory,
	}
	normalizerFactory := &loginid.NormalizerFactory{
		Config: loginIDConfig,
	}
	clockClock := _wireSystemClockValue
	provider := &loginid.Provider{
		Store:             loginidStore,
		Config:            loginIDConfig,
		Checker:           checker,
		NormalizerFactory: normalizerFactory,
		Clock:             clockClock,
	}
	oauthStore := &oauth.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	oauthProvider := &oauth.Provider{
		Store: oauthStore,
		Clock: clockClock,
	}
	anonymousStore := &anonymous.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	anonymousProvider := &anonymous.Provider{
		Store: anonymousStore,
		Clock: clockClock,
	}
	biometricStore := &biometric.Store{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	biometricProvider := &biometric.Provider{
		Store: biometricStore,
		Clock: clockClock,
	}
	serviceService := &service.Service{
		Authentication:        authenticationConfig,
		Identity:              identityConfig,
		IdentityFeatureConfig: identityFeatureConfig,
		Store:                 store,
		LoginID:               provider,
		OAuth:                 oauthProvider,
		Anonymous:             anonymousProvider,
		Biometric:             biometricProvider,
	}
	factory := appProvider.LoggerFactory
	logger := verification.NewLogger(factory)
	verificationConfig := appConfig.Verification
	rootProvider := appProvider.RootProvider
	environmentConfig := rootProvider.EnvironmentConfig
	trustProxy := environmentConfig.TrustProxy
	appredisHandle := appProvider.Redis
	storeRedis := &verification.StoreRedis{
		Redis: appredisHandle,
		AppID: appID,
		Clock: clockClock,
	}
	storePQ := &verification.StorePQ{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	ratelimitLogger := ratelimit.NewLogger(factory)
	storageRedis := &ratelimit.StorageRedis{
		AppID: appID,
		Redis: appredisHandle,
	}
	limiter := &ratelimit.Limiter{
		Logger:  ratelimitLogger,
		Storage: storageRedis,
		Clock:   clockClock,
	}
	verificationService := &verification.Service{
		Request:     request,
		Logger:      logger,
		Config:      verificationConfig,
		TrustProxy:  trustProxy,
		Clock:       clockClock,
		CodeStore:   storeRedis,
		ClaimStore:  storePQ,
		RateLimiter: limiter,
	}
	resolveHandlerLogger := handler.NewResolveHandlerLogger(factory)
	resolveHandler := &handler.ResolveHandler{
		Database:     handle,
		Identities:   serviceService,
		Verification: verificationService,
		Logger:       resolveHandlerLogger,
	}
	return resolveHandler
}

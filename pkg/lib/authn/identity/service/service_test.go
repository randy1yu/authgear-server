package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/google"
)

func newBool(b bool) *bool {
	return &b
}

func TestProviderListCandidates(t *testing.T) {
	Convey("Provider ListCandidates", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		loginIDProvider := NewMockLoginIDIdentityProvider(ctrl)
		oauthProvider := NewMockOAuthIdentityProvider(ctrl)
		siweProvider := NewMockSIWEIdentityProvider(ctrl)

		p := &Service{
			Authentication: &config.AuthenticationConfig{},
			Identity: &config.IdentityConfig{
				LoginID: &config.LoginIDConfig{},
				OAuth:   &config.OAuthSSOConfig{},
			},
			IdentityFeatureConfig: &config.IdentityFeatureConfig{
				OAuth: &config.OAuthSSOFeatureConfig{
					Providers: &config.OAuthSSOProvidersFeatureConfig{
						Google: &config.OAuthSSOProviderFeatureConfig{
							Disabled: false,
						},
					},
				},
			},
			LoginID: loginIDProvider,
			OAuth:   oauthProvider,
			SIWE:    siweProvider,
		}

		Convey("no candidates", func() {
			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldBeEmpty)
		})

		Convey("oauth", func() {
			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeOAuth}
			p.Identity.OAuth.Providers = []oauthrelyingparty.ProviderConfig{
				{
					"alias":           "google",
					"type":            google.Type,
					"modify_disabled": false,
				},
			}

			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":         "",
					"type":                "oauth",
					"display_id":          "",
					"provider_type":       "google",
					"provider_alias":      "google",
					"provider_subject_id": "",
					"provider_app_type":   "",
					"modify_disabled":     false,
				},
			})
		})

		Convey("loginid", func() {
			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeLoginID}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfig{
				{
					Type:           "email",
					Key:            "email",
					ModifyDisabled: newBool(false),
				},
			}

			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":     "",
					"type":            "login_id",
					"display_id":      "",
					"login_id_type":   "email",
					"login_id_key":    "email",
					"login_id_value":  "",
					"modify_disabled": false,
				},
			})
		})

		Convey("respect authentication", func() {
			p.Identity.OAuth.Providers = []oauthrelyingparty.ProviderConfig{
				{
					"alias":           "google",
					"type":            google.Type,
					"modify_disabled": false,
				},
			}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfig{
				{
					Type:           "email",
					Key:            "email",
					ModifyDisabled: newBool(false),
				},
			}

			actual, err := p.ListCandidates("")
			So(err, ShouldBeNil)
			So(actual, ShouldBeEmpty)
		})

		Convey("associate login ID identity", func() {
			userID := "a"

			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeLoginID}
			p.Identity.LoginID.Keys = []config.LoginIDKeyConfig{
				{
					Type:           "email",
					Key:            "email",
					ModifyDisabled: newBool(false),
				},
			}

			loginIDProvider.EXPECT().List(userID).Return([]*identity.LoginID{
				{
					LoginIDKey:      "email",
					LoginID:         "john.doe@example.com",
					OriginalLoginID: "john.doe@example.com",
					Claims: map[string]interface{}{
						"email": "john.doe@example.com",
					},
				},
			}, nil)
			oauthProvider.EXPECT().List(userID).Return(nil, nil)
			siweProvider.EXPECT().List(userID).Return(nil, nil)

			actual, err := p.ListCandidates(userID)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":     "",
					"type":            "login_id",
					"display_id":      "john.doe@example.com",
					"login_id_type":   "email",
					"login_id_key":    "email",
					"login_id_value":  "john.doe@example.com",
					"modify_disabled": false,
				},
			})
		})

		Convey("associate oauth identity", func() {
			userID := "a"

			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeOAuth}
			p.Identity.OAuth.Providers = []oauthrelyingparty.ProviderConfig{
				{
					"alias":           "google",
					"type":            google.Type,
					"modify_disabled": false,
				},
			}

			loginIDProvider.EXPECT().List(userID).Return(nil, nil)
			siweProvider.EXPECT().List(userID).Return(nil, nil)
			oauthProvider.EXPECT().List(userID).Return([]*identity.OAuth{
				{
					ProviderID: oauthrelyingparty.ProviderID{
						Type: google.Type,
						Keys: map[string]interface{}{},
					},
					ProviderSubjectID: "john.doe@gmail.com",
					Claims: map[string]interface{}{
						"email": "john.doe@gmail.com",
					},
				},
			}, nil)

			actual, err := p.ListCandidates(userID)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id":         "",
					"type":                "oauth",
					"display_id":          "john.doe@gmail.com",
					"provider_type":       "google",
					"provider_alias":      "google",
					"provider_subject_id": "john.doe@gmail.com",
					"provider_app_type":   "",
					"modify_disabled":     false,
				},
			})
		})

		Convey("associate siwe identity", func() {
			userID := "a"

			p.Authentication.Identities = []model.IdentityType{model.IdentityTypeSIWE}

			loginIDProvider.EXPECT().List(userID).Return(nil, nil)
			oauthProvider.EXPECT().List(userID).Return(nil, nil)
			siweProvider.EXPECT().List(userID).Return([]*identity.SIWE{
				{
					Address: "0x0",
				},
			}, nil)

			actual, err := p.ListCandidates(userID)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, []identity.Candidate{
				{
					"identity_id": "",
					"type":        "siwe",
					"display_id":  "0x0",
				},
			})
		})
	})
}

func TestProviderCheckDuplicated(t *testing.T) {
	Convey("Provider CheckDuplicated", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		loginIDProvider := NewMockLoginIDIdentityProvider(ctrl)
		oauthProvider := NewMockOAuthIdentityProvider(ctrl)

		p := &Service{
			Authentication: &config.AuthenticationConfig{},
			Identity: &config.IdentityConfig{
				LoginID: &config.LoginIDConfig{},
				OAuth:   &config.OAuthSSOConfig{},
			},
			IdentityFeatureConfig: &config.IdentityFeatureConfig{
				OAuth: &config.OAuthSSOFeatureConfig{
					Providers: &config.OAuthSSOProvidersFeatureConfig{
						Google: &config.OAuthSSOProviderFeatureConfig{
							Disabled: false,
						},
					},
				},
			},
			LoginID: loginIDProvider,
			OAuth:   oauthProvider,
		}

		makeEmailLoginID := func(userID string, email string) *identity.Info {
			return &identity.Info{
				UserID: userID,
				Type:   model.IdentityTypeLoginID,
				LoginID: &identity.LoginID{
					UserID:          userID,
					LoginIDKey:      "email",
					LoginIDType:     model.LoginIDKeyTypeEmail,
					LoginID:         email,
					OriginalLoginID: email,
					UniqueKey:       email,
					Claims: map[string]interface{}{
						"email": email,
					},
				},
			}
		}

		makeOAuth := func(userID string, providerSubjectID string, email string) *identity.Info {
			return &identity.Info{
				UserID: userID,
				Type:   model.IdentityTypeOAuth,
				OAuth: &identity.OAuth{
					UserID:            userID,
					ProviderSubjectID: providerSubjectID,
					Claims: map[string]interface{}{
						"email": email,
					},
				},
			}
		}

		Convey("brand new login ID", func() {
			info := makeEmailLoginID("user0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", info.LoginID.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().ListByClaim("email", info.LoginID.Claims["email"]).AnyTimes().Return(nil, nil)
			loginIDProvider.EXPECT().GetByUniqueKey(info.LoginID.UniqueKey).AnyTimes().Return(nil, identity.ErrIdentityNotFound)

			actual, err := p.CheckDuplicated(info)
			So(err, ShouldBeNil)
			So(actual, ShouldBeNil)
		})

		Convey("branch new oauth", func() {
			info := makeOAuth("user0", "google0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", info.OAuth.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().ListByClaim("email", info.OAuth.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().GetByProviderSubject(info.OAuth.ProviderID, info.OAuth.ProviderSubjectID).AnyTimes().Return(nil, identity.ErrIdentityNotFound)

			actual, err := p.CheckDuplicated(info)
			So(err, ShouldBeNil)
			So(actual, ShouldBeNil)
		})

		Convey("login ID / login ID clash; same user", func() {
			info := makeEmailLoginID("user0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", info.LoginID.Claims["email"]).AnyTimes().Return([]*identity.LoginID{info.LoginID}, nil)
			oauthProvider.EXPECT().ListByClaim("email", info.LoginID.Claims["email"]).AnyTimes().Return(nil, nil)
			loginIDProvider.EXPECT().GetByUniqueKey(info.LoginID.UniqueKey).AnyTimes().Return(info.LoginID, nil)

			actual, err := p.CheckDuplicated(info)
			So(err, ShouldBeNil)
			So(actual, ShouldBeNil)
		})

		Convey("login ID / login ID clash; different user", func() {
			incoming := makeEmailLoginID("user0", "johndoe@example.com")
			existing := makeEmailLoginID("user1", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", incoming.LoginID.Claims["email"]).AnyTimes().Return([]*identity.LoginID{existing.LoginID}, nil)
			oauthProvider.EXPECT().ListByClaim("email", incoming.LoginID.Claims["email"]).AnyTimes().Return(nil, nil)
			loginIDProvider.EXPECT().GetByUniqueKey(incoming.LoginID.UniqueKey).AnyTimes().Return(existing.LoginID, nil)

			actual, err := p.CheckDuplicated(incoming)
			So(errors.Is(err, identity.ErrIdentityAlreadyExists), ShouldBeTrue)
			So(actual, ShouldResemble, existing)
		})

		Convey("oauth / oauth clash; same user", func() {
			info := makeOAuth("user0", "google0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", info.OAuth.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().ListByClaim("email", info.OAuth.Claims["email"]).AnyTimes().Return([]*identity.OAuth{info.OAuth}, nil)
			oauthProvider.EXPECT().GetByProviderSubject(info.OAuth.ProviderID, info.OAuth.ProviderSubjectID).AnyTimes().Return(info.OAuth, nil)

			actual, err := p.CheckDuplicated(info)
			So(err, ShouldBeNil)
			So(actual, ShouldBeNil)
		})

		Convey("oauth / oauth clash; different user", func() {
			incoming := makeOAuth("user0", "google0", "johndoe@example.com")
			existing := makeOAuth("user1", "google0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", incoming.OAuth.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().ListByClaim("email", incoming.OAuth.Claims["email"]).AnyTimes().Return([]*identity.OAuth{existing.OAuth}, nil)
			oauthProvider.EXPECT().GetByProviderSubject(incoming.OAuth.ProviderID, incoming.OAuth.ProviderSubjectID).AnyTimes().Return(existing.OAuth, nil)

			actual, err := p.CheckDuplicated(incoming)
			So(errors.Is(err, identity.ErrIdentityAlreadyExists), ShouldBeTrue)
			So(actual, ShouldResemble, existing)
		})

		Convey("login / oauth clash; same user", func() {
			incoming := makeEmailLoginID("user0", "johndoe@example.com")
			existing := makeOAuth("user0", "google0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", incoming.LoginID.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().ListByClaim("email", incoming.LoginID.Claims["email"]).AnyTimes().Return([]*identity.OAuth{existing.OAuth}, nil)
			loginIDProvider.EXPECT().GetByUniqueKey(incoming.LoginID.UniqueKey).AnyTimes().Return(nil, identity.ErrIdentityNotFound)

			actual, err := p.CheckDuplicated(incoming)
			So(err, ShouldBeNil)
			So(actual, ShouldBeNil)
		})

		Convey("login / oauth clash; different user", func() {
			incoming := makeEmailLoginID("user0", "johndoe@example.com")
			existing := makeOAuth("user1", "google0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", incoming.LoginID.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().ListByClaim("email", incoming.LoginID.Claims["email"]).AnyTimes().Return([]*identity.OAuth{existing.OAuth}, nil)
			loginIDProvider.EXPECT().GetByUniqueKey(incoming.LoginID.UniqueKey).AnyTimes().Return(nil, identity.ErrIdentityNotFound)

			actual, err := p.CheckDuplicated(incoming)
			So(errors.Is(err, identity.ErrIdentityAlreadyExists), ShouldBeTrue)
			So(actual, ShouldResemble, existing)
		})

		Convey("oauth / login clash; same user", func() {
			incoming := makeOAuth("user0", "google0", "johndoe@example.com")
			existing := makeEmailLoginID("user0", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", incoming.OAuth.Claims["email"]).AnyTimes().Return([]*identity.LoginID{existing.LoginID}, nil)
			oauthProvider.EXPECT().ListByClaim("email", incoming.OAuth.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().GetByProviderSubject(incoming.OAuth.ProviderID, incoming.OAuth.ProviderSubjectID).AnyTimes().Return(nil, identity.ErrIdentityNotFound)

			actual, err := p.CheckDuplicated(incoming)
			So(err, ShouldBeNil)
			So(actual, ShouldBeNil)
		})

		Convey("oauth / login clash; different user", func() {
			incoming := makeOAuth("user0", "google0", "johndoe@example.com")
			existing := makeEmailLoginID("user1", "johndoe@example.com")

			loginIDProvider.EXPECT().ListByClaim("email", incoming.OAuth.Claims["email"]).AnyTimes().Return([]*identity.LoginID{existing.LoginID}, nil)
			oauthProvider.EXPECT().ListByClaim("email", incoming.OAuth.Claims["email"]).AnyTimes().Return(nil, nil)
			oauthProvider.EXPECT().GetByProviderSubject(incoming.OAuth.ProviderID, incoming.OAuth.ProviderSubjectID).AnyTimes().Return(nil, identity.ErrIdentityNotFound)

			actual, err := p.CheckDuplicated(incoming)
			So(errors.Is(err, identity.ErrIdentityAlreadyExists), ShouldBeTrue)
			So(actual, ShouldResemble, existing)
		})
	})
}

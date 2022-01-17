package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SettingsSubRoutesMiddlewareIdentityService interface {
	ListByUser(userID string) ([]*identity.Info, error)
}

// SettingsSubRoutesMiddleware redirect all settings sub routes to /settings
// if the current user is anonymous user
type SettingsSubRoutesMiddleware struct {
	Database   *appdb.Handle
	Identities SettingsSubRoutesMiddlewareIdentityService
}

func (m SettingsSubRoutesMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		// userID is nil should be blocked by RequireAuthenticatedMiddleware
		if userID == nil {
			next.ServeHTTP(w, r)
			return
		}

		isAnonymous := false
		err := m.Database.ReadOnly(func() (err error) {
			identities, err := m.Identities.ListByUser(*userID)
			if err != nil {
				return err
			}
			for _, i := range identities {
				if i.Type == model.IdentityTypeAnonymous {
					isAnonymous = true
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}

		if isAnonymous {
			http.Redirect(w, r, "/settings", http.StatusFound)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

package webapp

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSignupHTML = template.RegisterHTML(
	"web/signup.html",
	components...,
)

var SignupWithLoginIDSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_login_id_key": { "type": "string" },
			"x_login_id_type": { "type": "string" },
			"x_login_id_input_type": { "type": "string", "enum": ["email", "phone", "text"] },
			"x_login_id": { "type": "string" }
		},
		"required": ["x_login_id_key", "x_login_id_type", "x_login_id_input_type", "x_login_id"]
	}
`)

func ConfigureSignupRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern("/signup")
}

type MeterService interface {
	TrackPageView(VisitorID string, pageType meter.PageType) error
}

type SignupHandler struct {
	ControllerFactory       ControllerFactory
	BaseViewModel           *viewmodels.BaseViewModeler
	AuthenticationViewModel *viewmodels.AuthenticationViewModeler
	FormPrefiller           *FormPrefiller
	Renderer                Renderer
	MeterService            MeterService
	TutorialCookie          TutorialCookie
}

func (h *SignupHandler) GetData(r *http.Request, rw http.ResponseWriter, graph *interaction.Graph) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.EmbedForm(data, r.Form)
	if h.TutorialCookie.Pop(r, rw, httputil.SignupLoginTutorialCookieName) {
		baseViewModel.SetTutorial(httputil.SignupLoginTutorialCookieName)
	}
	viewmodels.Embed(data, baseViewModel)
	authenticationViewModel := h.AuthenticationViewModel.NewWithGraph(graph, r.Form)
	viewmodels.Embed(data, authenticationViewModel)
	return data, nil
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	h.FormPrefiller.Prefill(r.Form)

	opts := webapp.SessionOptions{
		RedirectURI: ctrl.RedirectURI(),
	}

	userIDHint := ""
	webhookState := ""
	suppressIDPSessionCookie := false
	prompt := []string{}
	if s := webapp.GetSession(r.Context()); s != nil {
		webhookState = s.WebhookState
		prompt = s.Prompt
		userIDHint = s.UserIDHint
		suppressIDPSessionCookie = s.SuppressIDPSessionCookie
	}
	intent := &intents.IntentAuthenticate{
		Kind:                     intents.IntentAuthenticateKindSignup,
		WebhookState:             webhookState,
		UserIDHint:               userIDHint,
		SuppressIDPSessionCookie: suppressIDPSessionCookie,
	}

	ctrl.Get(func() error {
		visitorID := webapp.GetVisitorID(r.Context())
		if visitorID == "" {
			// visitor id should be generated by VisitorIDMiddleware
			return fmt.Errorf("webapp: missing visitor id")
		}

		err := h.MeterService.TrackPageView(visitorID, meter.PageTypeSignup)
		if err != nil {
			return err
		}

		graph, err := ctrl.EntryPointGet(opts, intent)
		if err != nil {
			return err
		}

		data, err := h.GetData(r, w, graph)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSignupHTML, data)
		return nil
	})

	ctrl.PostAction("oauth", func() error {
		providerAlias := r.Form.Get("x_provider_alias")
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			input = &InputUseOAuth{
				ProviderAlias:    providerAlias,
				ErrorRedirectURI: httputil.HostRelative(r.URL).String(),
				Prompt:           prompt,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})

	ctrl.PostAction("login_id", func() error {
		result, err := ctrl.EntryPointPost(opts, intent, func() (input interface{}, err error) {
			err = SignupWithLoginIDSchema.Validator().ValidateValue(FormToJSON(r.Form))
			if err != nil {
				return
			}

			loginIDValue := r.Form.Get("x_login_id")
			loginIDKey := r.Form.Get("x_login_id_key")
			loginIDType := r.Form.Get("x_login_id_type")

			input = &InputNewLoginID{
				LoginIDType:  loginIDType,
				LoginIDKey:   loginIDKey,
				LoginIDValue: loginIDValue,
			}
			return
		})
		if err != nil {
			return err
		}

		result.WriteResponse(w, r)
		return nil
	})
}

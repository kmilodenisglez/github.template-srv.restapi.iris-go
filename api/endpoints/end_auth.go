package endpoints

import (
	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/hero"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/lib"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/repo/db"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema/dto"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/schema/mapper"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/service"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/service/auth"
	"github.com/kmilodenisglez/github.template-srv.restapi.iris.go/service/utils"
)

type HAuth struct {
	response  *utils.SvcResponse
	appConf   *utils.SvcConfig
	providers map[string]bool
	validate  *validator.Validate // handle validations for structs and individual fields based on tags
}

// NewAuthHandler create and register the authentication handlers for the App. For the moment, all the
// auth handlers emulates the Oauth2 "password" grant-type using the "client-credentials" flow.
//
// - app [*iris.Application] ~ Iris App instance
//
// - MdwAuthChecker [*context.Handler] ~ Authentication checker middleware
//
// - svcR [*utils.SvcResponse] ~ GrantIntentResponse service instance
//
// - svcC [utils.SvcConfig] ~ Configuration service instance
func NewAuthHandler(app *iris.Application, mdwAuthChecker *context.Handler, svcR *utils.SvcResponse, svcC *utils.SvcConfig, validate *validator.Validate) HAuth { // --- VARS SETUP ---
	h := HAuth{svcR, svcC, make(map[string]bool), validate}
	// filling providers
	h.providers["firstapp_provider"] = true

	repoDrones := db.NewRepoDrones(svcC)
	svcAuth := auth.NewSvcAuthentication(h.providers, &repoDrones) // instantiating authentication Service
	svcDrones := service.NewSvcDronesReqs(&repoDrones)

	// Simple group: v1
	v1 := app.Party("/api/v1")
	{
		// registering unprotected router
		authRouter := v1.Party("/auth") // authorize
		{
			// --- GROUP / PARTY MIDDLEWARES ---

			// --- DEPENDENCIES ---
			hero.Register(depObtainUserCred)
			hero.Register(svcAuth) // as an alternative, we can put these dependencies as property in the struct HAuth, as we are doing in the rest of the endpoints / handlers
			hero.Register(svcDrones)

			// --- REGISTERING ENDPOINTS ---
			// authRouter.Post("/<provider>")	// provider is the auth provider to be used.
			authRouter.Post("/", hero.Handler(h.authIntent))
		}

		// registering protected router
		guardAuthRouter := v1.Party("/auth")
		{
			// --- GROUP / PARTY MIDDLEWARES ---
			guardAuthRouter.Use(*mdwAuthChecker) // registering access token checker middleware

			// --- DEPENDENCIES ---
			hero.Register(DepObtainUserDid)
			hero.Register(repoDrones)

			// --- REGISTERING ENDPOINTS ---
			guardAuthRouter.Get("/logout", h.logout)
			guardAuthRouter.Get("/user", hero.Handler(h.userGet))
		}
	}
	return h
}

// region ======== ENDPOINT HANDLERS =====================================================

// authIntent Intent to grant authentication using the provider user's credentials and the specified  auth provider
// @Summary User authentication
// @description.markdown AuthIntent
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Param 	credential 	body 	dto.UserCredIn 	true	"User Login Credential"
// @Success 200 "OK"
// @Failure 401 {object} dto.Problem "err.unauthorized"
// @Failure 400 {object} dto.Problem "err.wrong_auth_provider"
// @Failure 504 {object} dto.Problem "err.network"
// @Failure 500 {object} dto.Problem "err.json_parse"
// @Router /auth [post]
func (h HAuth) authIntent(ctx iris.Context, uCred *dto.UserCredIn, svcAuth *auth.SvcAuthentication, r service.ISvcDrones) {
	// using a provider named 'drones', also injecting dependencies
	provider := "firstapp_provider"

	populate := r.IsPopulateDBSvc()
	if !populate {
		h.response.ResErr(&dto.Problem{Status: iris.StatusInternalServerError, Title: schema.ErrBuntdbNotPopulated, Detail: "The database has not been populated yet"}, &ctx)
		return
	}

	authGrantedData, problem := svcAuth.AuthProviders[provider].GrantIntent(uCred, nil) // requesting authorization to evote (provider) mechanisms in this case
	if problem != nil {                                                                 // check for errors
		h.response.ResErr(problem, &ctx)
		return
	}

	// TODO: pass this to the service
	// if so far so good, we are going to create the auth token
	tokenData := mapper.ToAccessTokenDataV(authGrantedData)
	accessToken, err := lib.MkAccessToken(tokenData, []byte(h.appConf.JWTSignKey), h.appConf.TkMaxAge)
	if err != nil {
		h.response.ResErr(&dto.Problem{Status: iris.StatusInternalServerError, Title: schema.ErrJwtGen, Detail: err.Error()}, &ctx)
		return
	}

	h.response.ResOKWithData(string(accessToken), &ctx)
}

// logout this endpoint invalidated a previously granted access token
// @Summary User logout
// @Description This endpoint invalidated a previously granted access token
// @Security ApiKeyAuth
// @Param Authorization header string true "Insert access token" default(Bearer <Add access token here>)
// @Tags Auth
// @Produce  json
// @Success 204 "OK"
// @Failure 401 {object} dto.Problem "err.unauthorized"
// @Failure 500 {object} dto.Problem "err.generic
// @Router /auth/logout [get]
func (h HAuth) logout(ctx iris.Context) {
	err := ctx.Logout()

	if err != nil {
		h.response.ResErr(&dto.Problem{Status: iris.StatusInternalServerError, Title: schema.ErrGeneric, Detail: err.Error()}, &ctx)
		return
	}

	// so far so good
	h.response.ResOK(&ctx)
}

// userGet Get user from the BD.
// @Security ApiKeyAuth
// @Param Authorization header string true "Insert access token" default(Bearer <Add access token here>)
// @Tags Auth
// @Produce  json
// @Success 200 {object} []dto.User "OK"
// @Failure 401 {object} dto.Problem "err.unauthorized"
// @Failure 500 {object} dto.Problem "err.generic
// @Router /auth/user [get]
func (h HAuth) userGet(ctx iris.Context, params dto.InjectedParam, r db.RepoDrones) {
	user, err := r.GetUser(params.Did, true)
	if err != nil {
		h.response.ResErr(dto.NewProblem(iris.StatusInternalServerError, schema.ErrBuntdb, err.Error()), &ctx)
		return
	}
	h.response.ResOKWithData(user, &ctx)
}

// endregion =============================================================================

// region ======== LOCAL DEPENDENCIES ====================================================

// depObtainUserCred is used as dependencies to obtain / create the user credential from request body (multipart/form-data).
// It returns a dto.UserCredIn struct
func depObtainUserCred(ctx iris.Context) dto.UserCredIn {
	cred := dto.UserCredIn{}

	// Getting data
	cred.Username = ctx.PostValue("username")
	cred.Password = ctx.PostValue("password")

	// TIP: We can do some validation here if we want
	return cred
}

// endregion =============================================================================

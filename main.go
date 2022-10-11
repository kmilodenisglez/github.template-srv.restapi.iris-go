package main

import (
	"fmt"
	ut "github.com/go-playground/universal-translator"
	enTranslations "github.com/go-playground/validator/v10/translations/en"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/validator/v10"
	"github.com/iris-contrib/swagger/v12"              // swagger middleware for Iris
	"github.com/iris-contrib/swagger/v12/swaggerFiles" // swagger embed files
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	_ "github.com/lib/pq"
	"restapi.app/api/endpoints"
	"restapi.app/api/middlewares"
	"restapi.app/docs"
	"restapi.app/lib"
	"restapi.app/service/cron"
	"restapi.app/service/utils"
)

func translations(validate *validator.Validate) ut.Translator {
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	_ = enTranslations.RegisterDefaultTranslations(validate, trans)
	return trans
}

func newApp() (*iris.Application, *utils.SvcConfig) {
	docs.SwaggerInfo.BasePath = "/api/v1"

	// region ======== GLOBALS ===============================================================
	validate := validator.New() // Validator instance. Reference https://github.com/kataras/iris/wiki/Model-validation | https://github.com/go-playground/validator
	trans := translations(validate)

	app := iris.New()        // App instance
	app.Validator = validate // Register validation on the iris app

	// Services
	svcConfig := utils.NewSvcConfig()              // Creating Configuration Service
	svcResponse := utils.NewSvcResponse(svcConfig) // Creating Response Service
	// endregion =============================================================================

	// region ======== MIDDLEWARES ===========================================================
	// Our custom CORS middleware.
	crs := func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Methods",
				"POST, PUT, PATCH, DELETE")

			ctx.Header("Access-Control-Allow-Headers",
				"Access-Control-Allow-Origin,Content-Type,authorization")

			ctx.Header("Access-Control-Max-Age",
				"86400")

			ctx.StatusCode(iris.StatusNoContent)
			return
		}

		ctx.Next()
	}

	// activate validator/v10 package and adding new validators
	err := lib.InitValidator(validate)
	if err != nil {
		panic(err.Error())
	}

	// built-ins
	app.Use(logger.New())
	app.UseRouter(crs) // Recovery middleware recovers from any panics and writes a 500 if there was one.

	// custom middleware
	mdwAuthChecker := middlewares.NewAuthCheckerMiddleware([]byte(svcConfig.JWTSignKey))

	// endregion =============================================================================

	// region ======== ENDPOINT REGISTRATIONS ================================================

	endpoints.NewAuthHandler(app, &mdwAuthChecker, svcResponse, svcConfig, validate)
	endpoints.NewFirstModuleHandler(app, &mdwAuthChecker, svcResponse, svcConfig, validate, trans) // Drones request handlers
	// endregion =============================================================================

	// region ======== SWAGGER REGISTRATION ==================================================
	// use swagger middleware to
	app.Get("/swagger/{any:path}", swagger.WrapHandler(swaggerFiles.Handler))
	// endregion =============================================================================

	return app, svcConfig
}

// @title GitHub template restapi
// @version 0.1
// @description REST API that allows clients to communicate with ... (i.e. **dispatch controller**)

// @contact.name Daniel Mena and Kmilo Denis Glez
// @contact.url https://github.com/dani-fmena and https://github.com/kmilodenisglez
// @contact.email kmilo.denis.glez@gmail.com

// @authorizationurl https://example.com/oauth/authorize

// TIPS This Ip here 👇🏽  must be change when compiling to deploy, can't figure out how to do it dynamically with Iris.

// @BasePath /
func main() {
	app, svcConfig := newApp()

	// region ======== Cron Job ==================================================
	cronJob := cron.NewSvcRepoEventLog(svcConfig)
	_ = cronJob.MeinerCronJob()
	// endregion =============================================================================

	addr := fmt.Sprintf(":%s", svcConfig.DappPort)

	app.Run(iris.Addr(addr))
}

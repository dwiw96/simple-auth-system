package delivery

import (
	"context"
	"net/http"

	auth "github.com/dwiw96/simple-auth-system/features/auth"
	mid "github.com/dwiw96/simple-auth-system/middleware"
	responses "github.com/dwiw96/simple-auth-system/utils/responses"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type authDelivery struct {
	router   *gin.Engine
	service  auth.ServiceInterface
	validate *validator.Validate
	trans    ut.Translator
}

func NewAuthDelivery(router *gin.Engine, service auth.ServiceInterface, pool *pgxpool.Pool, client *redis.Client, ctx context.Context) {
	handler := &authDelivery{
		router:   router,
		service:  service,
		validate: validator.New(),
	}

	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(handler.validate, trans)
	handler.trans = trans

	router.POST("/api/signup", handler.SignUp)
	router.POST("/api/login", handler.LogIn)
	router.POST("/api/logout", mid.AuthMiddleware(ctx, pool, client), handler.LogOut)
}

func translateError(trans ut.Translator, err error) (errTrans []string) {
	errs := err.(validator.ValidationErrors)
	a := (errs.Translate(trans))
	for _, val := range a {
		errTrans = append(errTrans, val)
	}

	return
}

func (d *authDelivery) SignUp(c *gin.Context) {
	var request signupRequest

	err := c.BindJSON(&request)

	if err != nil {
		c.JSON(422, err.Error())
		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	err = d.validate.Struct(request)
	if err != nil {
		errTranslated := translateError(d.trans, err)
		responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
		return
	}

	signupInput := toSignUpRequest(request)
	user, code, err := d.service.SignUp(signupInput)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toSignUpResponse(user)

	response := responses.SuccessWithDataResponse(respBody, 201, "Sign up success")
	c.IndentedJSON(http.StatusCreated, response)
}

func (d *authDelivery) LogIn(c *gin.Context) {
	var request signinRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(422, err)
		return
	}

	err = d.validate.Struct(request)
	if err != nil {
		errTranslated := translateError(d.trans, err)
		responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
		return
	}

	user, token, code, err := d.service.LogIn(auth.LoginRequest(request))
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toSignUpResponse(user)

	response := responses.SuccessWithDataResponse(respBody, 200, "Login success")
	c.Header("Authorization", token)
	c.IndentedJSON(200, response)
}

func (d *authDelivery) LogOut(c *gin.Context) {
	authPayload, isExists := c.Keys["payloadKey"].(*auth.JwtPayload)

	if !isExists {
		responses.ErrorJSON(c, 401, []string{"token is wrong"}, c.Request.RemoteAddr)
		c.JSON(401, "token is wrong")
		return
	}

	err := d.service.LogOut(*authPayload)
	if err != nil {
		responses.ErrorJSON(c, 401, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	c.JSON(200, "logout success")
}

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

	router.Use(mid.AuthMiddleware(ctx, pool, client))

	router.POST("/api/signup", handler.signUp)
	router.POST("/api/login", handler.logIn)
	router.POST("/api/logout", handler.logOut)
	router.PUT("/api/email/verification", handler.emailVerification)
	router.PUT("/api/email/unverification", handler.emailUnVerification)
}

func translateError(trans ut.Translator, err error) (errTrans []string) {
	errs := err.(validator.ValidationErrors)
	a := (errs.Translate(trans))
	for _, val := range a {
		errTrans = append(errTrans, val)
	}

	return
}

func (d *authDelivery) signUp(c *gin.Context) {
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

func (d *authDelivery) logIn(c *gin.Context) {
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

func (d *authDelivery) logOut(c *gin.Context) {
	authPayload, isExists := c.Keys["payloadKey"].(*auth.JwtPayload)

	if !isExists {
		responses.ErrorJSON(c, 401, []string{"token is wrong"}, c.Request.RemoteAddr)
		return
	}

	err := d.service.LogOut(*authPayload)
	if err != nil {
		responses.ErrorJSON(c, 401, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	c.JSON(200, "logout success")
}

func (d *authDelivery) emailVerification(c *gin.Context) {
	authPayload, isExists := c.Keys["payloadKey"].(*auth.JwtPayload)
	if !isExists {
		responses.ErrorJSON(c, 401, []string{"token is wrong"}, c.Request.RemoteAddr)
		c.JSON(401, "token is wrong")
		return
	}

	code, err := d.service.EmailVerification(authPayload.UserID, authPayload.Email)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessResponse("email is verified")
	c.IndentedJSON(200, response)
}

func (d *authDelivery) emailUnVerification(c *gin.Context) {
	authPayload, isExists := c.Keys["payloadKey"].(*auth.JwtPayload)
	if !isExists {
		responses.ErrorJSON(c, 401, []string{"token is wrong"}, c.Request.RemoteAddr)
		c.JSON(401, "token is wrong")
		return
	}

	code, err := d.service.DeleteUser(authPayload.UserID, authPayload.Email)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessResponse("user is unverify and deleted")
	c.IndentedJSON(200, response)
}

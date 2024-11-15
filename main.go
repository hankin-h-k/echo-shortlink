package main

import (
	"net/http"
	"shorturl/env"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
)

type ShortlinkReq struct {
	Url                 string `json:"url" form:"url" validate:"required"`
	ExpirationInMinutes int64  `json:"expiration_in_munutes" form:"expiration_in_munutes" validate:"required"`
}
type ShortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

type CustomValidator struct {
	validator *validator.Validate
}

var conf *env.Env

func main() {
	conf = env.GetEnv()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	api := e.Group("/api")
	{
		api.POST("/shorten", createShortlink)
		api.GET("/info", getShortlinkInfo)
		api.GET("/shortlink/:id", redirect)
	}
	e.Logger.Fatal(e.Start(":8080"))
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	c.JSON(http.StatusOK, err.Error())
}

func createShortlink(c echo.Context) (err error) {
	var req ShortlinkReq
	if err = c.Bind(&req); err != nil {
		return
	}
	if err = c.Validate(req); err != nil {
		return
	}
	eid, err := conf.S.Shorten(req.Url, req.ExpirationInMinutes)
	if err != nil {
		log.Error(err)
	}
	return c.JSON(http.StatusOK, eid)
}

func getShortlinkInfo(c echo.Context) (err error) {
	eid := c.QueryParam("eid")
	res, err := conf.S.ShortlinkInfo(eid)
	if err != nil {
		log.Error(err)
	}
	return c.JSON(http.StatusOK, res)
}

func redirect(c echo.Context) (err error) {
	eid := c.Param("id")
	url, err := conf.S.Unshorten(eid)
	if err != nil {
		log.Error(err)
	}
	return c.Redirect(http.StatusFound, url)
}

package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})

	e.GET("/me", func(c echo.Context) error {
		fmt.Println(c.Request().Header)
		username := c.Request().Header.Get("X-Auth-User")
		userID := c.Request().Header.Get("X-Auth-User-ID")
		return c.JSON(http.StatusOK, map[string]string{
			"message":  "Hello, World!",
			"username": username,
			"userID":   userID,
		})
	})

	e.Logger.Fatal(e.Start(":8002"))
}

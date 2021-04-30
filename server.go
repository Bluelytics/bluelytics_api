package main

import (
	"net/http"
	
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORS())

	e.GET("/", index)

	db := dbPool()

	e.GET("v2/latest", func (c echo.Context) error {
		return v2_latest(c, db)
	})
	e.GET("v2/alerts", func (c echo.Context) error {
		return v2_alerts(c, db)
	})
	e.GET("json/last_price", func (c echo.Context) error {
		return v1_json_last_price(c, db)
	})
	e.GET("data/graphs/evolution.json", func (c echo.Context) error {
		return dolar_evolution(c, db)
	})
	e.GET("data/json/currency.json", func (c echo.Context) error {
		return currency(c, db)
	})


	e.Logger.Fatal(e.Start(":8080"))
}

func index(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
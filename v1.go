package main
import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type v1_json_dolar struct {
	Date time.Time `json:"date"`
	Source string `json:"source"`
	Value_avg Number `json:"value_avg"`
	Value_sell Number `json:"value_sell"`
	Value_buy Number `json:"value_buy"`
}

func v1_json_last_price(c echo.Context, db *pgxpool.Pool) error {
	dolar := getDolarData(db)

	var json_v1 [2]v1_json_dolar
	json_v1[0] = v1_json_dolar{dolar.Last_update, "oficial", dolar.Oficial.Value_avg, dolar.Oficial.Value_sell, dolar.Oficial.Value_buy}
	json_v1[1] = v1_json_dolar{dolar.Last_update, "blue", dolar.Oficial.Value_avg, dolar.Oficial.Value_sell, dolar.Oficial.Value_buy}


	callback := c.QueryParam("callback")

	if(callback != ""){
		return c.JSONP(http.StatusOK, callback, json_v1)
	}else{
		return c.JSON(http.StatusOK, json_v1)
	}
}

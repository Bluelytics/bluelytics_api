package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
)

type Number float64

func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%.2f", n)), nil
}

func dbPool() *pgxpool.Pool {
	connStr := fmt.Sprintf("user=%s dbname=%s pool_min_conns=5 pool_max_conns=15", os.Getenv("PG_USER"), os.Getenv("PG_DB"))
	ssl_mode := os.Getenv("PG_SSLMODE")
	password := os.Getenv("PG_PASSWORD")
	hostname := os.Getenv("PG_HOSTNAME")

	if ssl_mode != "" {
		connStr = connStr + fmt.Sprintf(" sslmode=%s", ssl_mode)
	}
	if password != "" {
		connStr = connStr + fmt.Sprintf(" password=%s", password)
	}
	if hostname != "" {
		connStr = connStr + fmt.Sprintf(" host=%s", hostname)
	}

	pgConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatal(err)
	}

	return pool
}

type MonedaValues struct {
	Value_avg  Number `json:"value_avg"`
	Value_sell Number `json:"value_sell"`
	Value_buy  Number `json:"value_buy"`
}

type Dolares struct {
	Oficial      MonedaValues `json:"oficial"`
	Blue         MonedaValues `json:"blue"`
	Oficial_euro MonedaValues `json:"oficial_euro"`
	Blue_euro    MonedaValues `json:"blue_euro"`
	Last_update  time.Time    `json:"last_update"`
}

func getDolarData(db *pgxpool.Pool) Dolares {
	var res Dolares
	rows, err := db.Query(context.Background(), `
		select tipo, updated_at, value_buy, value_sell, euro_buy, euro_sell
	from dolar_latest
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var updated_at time.Time
		var dolar_buy Number
		var dolar_sell Number
		var euro_buy Number
		var euro_sell Number

		if err := rows.Scan(&name, &updated_at, &dolar_buy, &dolar_sell, &euro_buy, &euro_sell); err != nil {
			log.Fatal(err)
		}

		if name == "oficial" {
			res.Oficial.Value_buy = dolar_buy
			res.Oficial.Value_sell = dolar_sell
			res.Oficial.Value_avg = (dolar_sell + dolar_buy) / 2
			res.Oficial_euro.Value_buy = euro_buy
			res.Oficial_euro.Value_sell = euro_sell
			res.Oficial_euro.Value_avg = (euro_sell + euro_buy) / 2
		} else {
			tz, err := time.LoadLocation("America/Buenos_Aires")
			if err != nil {
				log.Fatal(err)
			}
			res.Last_update = updated_at.In(tz)
			res.Blue.Value_buy = dolar_buy
			res.Blue.Value_sell = dolar_sell
			res.Blue.Value_avg = (dolar_sell + dolar_buy) / 2
			res.Blue_euro.Value_buy = euro_buy
			res.Blue_euro.Value_sell = euro_sell
			res.Blue_euro.Value_avg = (euro_sell + euro_buy) / 2
		}
	}

	return res
}

func v2_latest(c echo.Context, db *pgxpool.Pool) error {
	dolar := getDolarData(db)

	callback := c.QueryParam("callback")

	if callback != "" {
		return c.JSONP(http.StatusOK, callback, dolar)
	} else {
		return c.JSON(http.StatusOK, dolar)
	}

}

type Alertas struct {
	Id       uint32       `json:"alert_id"`
	Alert_at time.Time    `json:"alert_at"`
	Current  MonedaValues `json:"current"`
	Previous MonedaValues `json:"previous"`
}

func getAlertsData(db *pgxpool.Pool) Alertas {
	var res Alertas
	rows, err := db.Query(context.Background(), `
	with ranked as (
		select
		id,
		alert_at,
		value_buy,
		value_sell,
		(value_buy+value_sell)/2 as value_avg,
		lag(value_buy) over (order by alert_at) as last_value_sell,
		lag(value_sell) over (order by alert_at) as last_value_buy,
		lag((value_buy+value_sell)/2) over (order by alert_at) as last_value_avg,
		row_number() over(order by alert_at desc) rnk
		from
		alertas_dolar
		)
		select id, alert_at, value_buy, value_sell, value_avg, last_value_sell, last_value_buy, last_value_avg
		from ranked
		where rnk = 1
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {

		err := rows.Scan(
			&res.Id,
			&res.Alert_at,
			&res.Current.Value_buy,
			&res.Current.Value_sell,
			&res.Current.Value_avg,
			&res.Previous.Value_buy,
			&res.Previous.Value_sell,
			&res.Previous.Value_avg)
		if err != nil {
			log.Fatal(err)
		}
	}

	return res
}

func v2_alerts(c echo.Context, db *pgxpool.Pool) error {
	alert := getAlertsData(db)

	return c.JSON(http.StatusOK, alert)
}

// Dolares historical

type DolaresHistorical struct {
	Oficial MonedaValues `json:"oficial"`
	Blue    MonedaValues `json:"blue"`
}

type ErrorJson struct {
	Error string `json:"error"`
}

func getHistoricalData(day string, db *pgxpool.Pool) (DolaresHistorical, error) {
	var res DolaresHistorical

	date, err := time.Parse("2006-01-02", day)
	if err != nil {
		return res, fmt.Errorf("parsing day parameter: use YYYY-MM-DD")
	}

	if date.Weekday() == time.Saturday {
		// Use friday's data
		date = date.AddDate(0, 0, -1)
	}

	if date.Weekday() == time.Sunday {
		// Use friday's data
		date = date.AddDate(0, 0, -2)
	}

	rows, err := db.Query(context.Background(), `
	select
		lower(tipo) as tipo, value_sell, value_buy
	from
		dolar_evolution
	where dttm = $1
	`, date)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var dolar_buy Number
		var dolar_sell Number

		if err := rows.Scan(&name, &dolar_buy, &dolar_sell); err != nil {
			log.Fatal(err)
		}

		if name == "oficial" {
			res.Oficial.Value_buy = dolar_buy
			res.Oficial.Value_sell = dolar_sell
			res.Oficial.Value_avg = (dolar_sell + dolar_buy) / 2
		} else {
			res.Blue.Value_buy = dolar_buy
			res.Blue.Value_sell = dolar_sell
			res.Blue.Value_avg = (dolar_sell + dolar_buy) / 2
		}
	}

	if res.Blue.Value_sell == 0 {
		return res, fmt.Errorf("day not found")
	}

	return res, nil
}

func v2_historical(c echo.Context, db *pgxpool.Pool) error {
	day := c.QueryParam("day")
	dolarhistorico, err := getHistoricalData(day, db)
	if err != nil {
		var errjson ErrorJson
		errjson.Error = fmt.Sprintf("Error: %s", err.Error())

		return c.JSON(http.StatusNotFound, errjson)
	}

	return c.JSON(http.StatusOK, dolarhistorico)
}

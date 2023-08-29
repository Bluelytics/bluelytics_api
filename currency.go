package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
)

type CurrencyValue struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Value Number `json:"value"`
}

func getCurrencydata(db *pgxpool.Pool) []CurrencyValue {
	var res []CurrencyValue
	rows, err := db.Query(context.Background(), `
		select code, name, value
	from currency_value_latest
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var objAppend CurrencyValue

		if err := rows.Scan(&objAppend.Code, &objAppend.Name, &objAppend.Value); err != nil {
			log.Fatal(err)
		}

		res = append(res, objAppend)
	}

	return res
}

func currency(c echo.Context, db *pgxpool.Pool) error {
	res_data := getCurrencydata(db)

	return c.JSON(http.StatusOK, res_data)
}

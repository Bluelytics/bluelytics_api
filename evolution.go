package main
import (
	"context"
	"net/http"
	"log"
	"time"
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type DolarEvolutionDay struct {
	Date string `json:"date"`
	Source string `json:"source"`
	Value Number `json:"value"`
}

func getDolarEvolutionData(days string, db *pgxpool.Pool) ([]DolarEvolutionDay) {
	limitStr := ""
	var res []DolarEvolutionDay
	if days != "" {
		days_num, err := strconv.Atoi(days)
		if err == nil && days_num < 10000 {
			limitStr = "limit " + days
		}
	}
	

	rows,err := db.Query(context.Background(), `
		select dttm, tipo, value_sell
	from dolar_evolution
	order by dttm desc
	` + limitStr)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next(){
		var objAppend DolarEvolutionDay
		var dateObj time.Time


		if err := rows.Scan(&dateObj, &objAppend.Source, &objAppend.Value); err != nil {
			log.Fatal(err)
		}
		objAppend.Date = fmt.Sprintf(dateObj.Format("2006-01-02"))

		res = append(res, objAppend)
	}

	return res
}

func dolar_evolution(c echo.Context, db *pgxpool.Pool) error {
	days := c.QueryParam("days")
	res_data := getDolarEvolutionData(days, db)

	return c.JSON(http.StatusOK, res_data)
}

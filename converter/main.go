package main

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Result exported
type Result struct {
	Water    float64
	Sediment float64
	Tp       float64
	Tn       float64
}

var subbasinArray = make(map[int]map[int]Result)

var fieldArray = make(map[int]map[int]*Result)

var subbasinArea = make(map[int]float64)

var fieldArea = make(map[int]float64)

var subbasinAverage = make(map[int]*Result)
var fieldAverage = make(map[int]*Result)

func main() {
	db, err := sql.Open("sqlite3", "result.db3")
	checkErr(err)

	//查询数据
	rows, err := db.Query("SELECT id,year,water,sediment,TP,TN FROM sub ")
	checkErr(err)

	for rows.Next() {
		var id int
		var year int
		var water float64
		var sediment float64
		var tp float64
		var tn float64
		err = rows.Scan(&id, &year, &water, &sediment, &tp, &tn)
		checkErr(err)

		if len(subbasinArray[id]) == 0 {
			subbasinArray[id] = make(map[int]Result)
		}

		subbasinArray[id][year] = Result{water, sediment, tp, tn}
	}

	for k := range subbasinArray {
		fmt.Println(len(subbasinArray[k]))
		fmt.Println(subbasinArray[k])
	}

	db.Close()

	dbSpatial, err := sql.Open("sqlite3", "spatial.db3")

	subbasinAreaSearch, err := dbSpatial.Query("SELECT id,area FROM subbasin_area ")
	checkErr(err)

	for subbasinAreaSearch.Next() {
		var id int
		var area float64

		err = subbasinAreaSearch.Scan(&id, &area)
		checkErr(err)

		subbasinArea[id] = area
	}

	fieldAreaSearch, err := dbSpatial.Query("SELECT id,area FROM field_area ")
	checkErr(err)

	for fieldAreaSearch.Next() {
		var id int
		var area float64

		err = fieldAreaSearch.Scan(&id, &area)
		checkErr(err)

		fieldArea[id] = area
	}

	//查询数据
	rowsSpatial, err := dbSpatial.Query("SELECT field,subbasin,percent FROM field_subbasin ")
	checkErr(err)

	for rowsSpatial.Next() {
		var field int
		var subbasin int
		var percentage float64

		err = rowsSpatial.Scan(&field, &subbasin, &percentage)
		checkErr(err)

		if len(fieldArray[field]) == 0 {
			fieldArray[field] = make(map[int]*Result)
		}

		for i := 2002; i < 2012; i++ {

			if fieldArray[field][i] == nil {
				fieldArray[field][i] = new(Result)
			}

			fieldArray[field][i].Water = subbasinArray[subbasin][i].Water*percentage*subbasinArea[subbasin]/fieldArea[field] + fieldArray[field][i].Water
			fieldArray[field][i].Sediment = subbasinArray[subbasin][i].Sediment*percentage*subbasinArea[subbasin]/fieldArea[field] + fieldArray[field][i].Sediment
			fieldArray[field][i].Tp = subbasinArray[subbasin][i].Tp*percentage*subbasinArea[subbasin]/fieldArea[field] + fieldArray[field][i].Tp
			fieldArray[field][i].Tn = subbasinArray[subbasin][i].Tn*percentage*subbasinArea[subbasin]/fieldArea[field] + fieldArray[field][i].Tn
		}
	}

	for k := range fieldArray {
		fmt.Println(len(fieldArray[k]))
		fmt.Println("filed" + strconv.Itoa(k))

		for m := range fieldArray[k] {
			fmt.Println(*fieldArray[k][m])

		}
	}
	dbSpatial.Close()

	fmt.Println("the length of subbasinArray is " + strconv.Itoa(len(subbasinArray)))
	for k := range subbasinArray {
		var result = new(Result)

		for m := range subbasinArray[k] {
			result.Water = subbasinArray[k][m].Water + result.Water
			result.Sediment = subbasinArray[k][m].Sediment + result.Sediment
			result.Tn = subbasinArray[k][m].Tn + result.Tn
			result.Tp = subbasinArray[k][m].Tp + result.Tp
		}

		subbasinAverage[k] = new(Result)

		subbasinAverage[k].Sediment = result.Sediment / 10
		subbasinAverage[k].Water = result.Water / 10
		subbasinAverage[k].Tn = result.Tn / 10
		subbasinAverage[k].Tp = result.Tp / 10
		fmt.Println("subbasin" + strconv.Itoa(k))
		fmt.Println(subbasinAverage[k])

	}

	fmt.Println("the length of fieldArray is " + strconv.Itoa(len(fieldArray)))

	for k := range fieldArray {
		var result = new(Result)

		for m := range fieldArray[k] {
			result.Water = fieldArray[k][m].Water + result.Water
			result.Sediment = fieldArray[k][m].Sediment + result.Sediment
			result.Tn = fieldArray[k][m].Tn + result.Tn
			result.Tp = fieldArray[k][m].Tp + result.Tp
		}

		fieldAverage[k] = new(Result)

		fieldAverage[k].Sediment = result.Sediment / 10
		fieldAverage[k].Water = result.Water / 10
		fieldAverage[k].Tn = result.Tn / 10
		fieldAverage[k].Tp = result.Tp / 10
		fmt.Println("filed" + strconv.Itoa(k))
		fmt.Println(fieldAverage[k])

	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

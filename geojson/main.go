package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/montanaflynn/stats"
)

type Properties struct {
	Name string `json:"name"`
}

type Crs struct {
	Type       string      `json:"type"`
	Properties *Properties `json:"properties"`
}

type FeatureProperties struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Sediment      float64 `json:"sediment"`
	Flow          float64 `json:"flow"`
	Tp            float64 `json:"tp"`
	Tn            float64 `json:"tn"`
	SedimentLevel string  `json:"sedimentlevel"`
	FlowLevel     string  `json:"flowlevel"`
	TpLevel       string  `json:"tplevel"`
	TnLevel       string  `json:"tnlevel"`
}

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

type Feature struct {
	Type       string             `json:"type"`
	Properties *FeatureProperties `json:"properties"`
	Geometry   *Geometry          `json:"geometry"`
}

type MapFeature struct {
	Type     string    `json:"type"`
	Crs      *Crs      `json:"crs"`
	Features []Feature `json:"features"`
}

type Result struct {
	Water    float64
	Sediment float64
	Tp       float64
	Tn       float64
}

var SedimentQuartile []float64
var FlowQuartile []float64
var TpQuartile []float64
var TnQuartile []float64

var subbasinArray = make(map[int]map[int]Result)

var fieldArray = make(map[int]map[int]*Result)

var subbasinArea = make(map[int]float64)

var fieldArea = make(map[int]float64)

var subbasinAverage = make(map[int]*Result)

var fieldAverage = make(map[int]*Result)

func main() {
	var featureCollection = new(MapFeature)
	configFile, err := os.Open("field.json")
	if err != nil {
		fmt.Println("fail 1")
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&featureCollection); err != nil {
		fmt.Println("fail 2")
	}

	BasintoField()
	Quartile(fieldAverage)
	Quartile(subbasinAverage)

	// var temp int

	for id := range fieldAverage {
		for i := 0; i < len(featureCollection.Features); i++ {
			if strconv.Itoa(id) == featureCollection.Features[i].Properties.Name {
				featureCollection.Features[i].Properties.Flow = fieldAverage[id].Water
				featureCollection.Features[i].Properties.Sediment = fieldAverage[id].Sediment
				featureCollection.Features[i].Properties.Tp = fieldAverage[id].Tp
				featureCollection.Features[i].Properties.Tn = fieldAverage[id].Tn
				featureCollection.Features[i].Properties.FlowLevel = SelectLevel(fieldAverage[id].Water, FlowQuartile)
				featureCollection.Features[i].Properties.SedimentLevel = SelectLevel(fieldAverage[id].Sediment, SedimentQuartile)
				featureCollection.Features[i].Properties.TpLevel = SelectLevel(fieldAverage[id].Tp, TpQuartile)
				featureCollection.Features[i].Properties.TnLevel = SelectLevel(fieldAverage[id].Tn, TnQuartile)
			}
		}
	}

	b, err := json.MarshalIndent(featureCollection, "", "  ")
	// var i = len(featureCollection.Features)
	err = ioutil.WriteFile("./geojson/output.json", b, 0644)
	err = ioutil.WriteFile("output.json", b, 0644)

	if err != nil {
		panic(err)
	}

	fmt.Println("success")
	return
}

func BasintoField() {
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
		// fmt.Println("subbasin" + strconv.Itoa(k))
		// fmt.Println(subbasinAverage[k])

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
	}

}

func Quartile(m map[int]*Result) {
	var flowArray []float64
	var sedimentArray []float64
	var tnArray []float64
	var tpArray []float64

	for i := range m {
		sedimentArray = append(sedimentArray, m[i].Sediment)
		flowArray = append(flowArray, m[i].Water)
		tnArray = append(tnArray, m[i].Tn)
		tpArray = append(tpArray, m[i].Tp)
	}
	q, _ := stats.Quartile(sedimentArray)

	a1, _ := stats.Percentile(flowArray, 10)
	b1, _ := stats.Percentile(flowArray, 30)
	c1, _ := stats.Percentile(flowArray, 50)
	d1, _ := stats.Percentile(flowArray, 70)
	e1, _ := stats.Percentile(flowArray, 90)

	FlowQuartile = []float64{a1, b1, c1, d1, e1}

	a2, _ := stats.Percentile(sedimentArray, 10)
	b2, _ := stats.Percentile(sedimentArray, 30)
	c2, _ := stats.Percentile(sedimentArray, 50)
	d2, _ := stats.Percentile(sedimentArray, 70)
	e2, _ := stats.Percentile(sedimentArray, 90)
	SedimentQuartile = []float64{a2, b2, c2, d2, e2}

	a3, _ := stats.Percentile(tnArray, 10)
	b3, _ := stats.Percentile(tnArray, 30)
	c3, _ := stats.Percentile(tnArray, 50)
	d3, _ := stats.Percentile(tnArray, 70)
	e3, _ := stats.Percentile(tnArray, 90)
	TnQuartile = []float64{a3, b3, c3, d3, e3}

	a4, _ := stats.Percentile(tpArray, 10)
	b4, _ := stats.Percentile(tpArray, 30)
	c4, _ := stats.Percentile(tpArray, 50)
	d4, _ := stats.Percentile(tpArray, 70)
	e4, _ := stats.Percentile(tpArray, 90)

	TpQuartile = []float64{a4, b4, c4, d4, e4}

	fmt.Println(FlowQuartile)
	fmt.Println(SedimentQuartile)
	fmt.Println(TnQuartile)
	fmt.Println(TpQuartile) // 4
	fmt.Println(q)          // {15 37.5 40}}
}

var Level = []string{"Great", "Good", "Normal", "Slight", "Bad", "Severe"}

func SelectLevel(v float64, x []float64) (s string) {

	var found bool
	for i := range x {
		if v < x[i] {
			s = Level[i]
			found = true
			return s
		}
	}
	if found != true {
		return Level[len(x)]
	}
	return "Not found"
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

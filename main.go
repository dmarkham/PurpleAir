package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type PurpleDate struct {
	Fields []string
	Data   [][]interface{}
}

type output struct {
	Version  string `json:"version,omitempty"`
	FullText string `json:"full_text,omitempty"`
	MinWidth string `json:"min_width,omitempty"`
	Color    string `json:"color,omitempty"`
}

func main() {

	id := flag.String("ids", "", "ID's to Avg Pipe delimited")
	key := flag.String("key", "", "Key from a widget sample")
	flag.Parse()

	if *id == "" {
		panic("Missing Id's")
	}
	if *key == "" {
		panic("Missing key")
	}
	apiURL := "https://www.purpleair.com/data.json?key=" + *key + "&show=" + *id
	resp, err := http.Get(apiURL)
	if err != nil {
		panic(errors.WithStack(err))
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(errors.WithStack(err))
	}
	defer resp.Body.Close()
	var pData PurpleDate

	err = json.Unmarshal([]byte(data), &pData)
	if err != nil {
		panic(errors.WithStack(err))
	}
	//fmt.Println(pData)
	// ["ID","pm","age","pm_0","pm_1","pm_2","pm_3","pm_4","pm_5","pm_6",
	// "conf","pm1","pm_10","p1","p2","p3","p4","p5","p6","Humidity",
	// "Temperature","Pressure","Elevation","Type","Label","Lat","Lon","Icon","isOwner","Flags",
	// "Voc","Ozone1","Adc","CH"]

	// [35307,27.5,0,27.5,27.3,27.3,26.5,20.1,16.5,13.8,
	// 100,15.2,30.9,2765.3,806.4,207.3,23.3,3.3,2.1,58,
	// 81,1004.94,65,0,"2819 W CANYON AVE",32.798355,-117.11712,0,0,0,
	// null,null,0.0,3]

	sumPM := float64(0)
	sumTemp := float64(0)
	count := float64(0)
	for _, row := range pData.Data {
		//stationID, ok := row[0].(float64)
		//if !ok {
		//	panic(errors.New("Station ID did not Convert"))
		//	}

		//humidity, ok := row[19].(float64)
		//if !ok {
		//	panic(errors.New("Humidity did not Convert"))
		//	}
		pm, ok := row[1].(float64)
		if !ok {
			panic(errors.New("PM did not Convert"))
		}
		temp, ok := row[20].(float64)
		if !ok {
			panic(errors.New("Temp did not Convert"))
		}
		temp = temp - 8

		//fmt.Printf("StationID:%v, PM2.5:%v Temp: %v, Humidity: %v\n", stationID, pm, temp, humidity)
		count++
		sumPM = sumPM + pm
		sumTemp = sumTemp + temp
	}

	pmavg := sumPM / count
	color := "#00ff4c"
	if pmavg > 12 {
		color = "#c78500"
	} else if pmavg > 35 {
		color = "#c91a1a"
	}

	o := output{
		Version:  "1",
		FullText: fmt.Sprintf("Air Quality: %0.1f Temp: %0.1f ", pmavg, sumTemp/count),
		Color:    color,
	}
	bytes, err := json.Marshal(o)
	if err != nil {
		panic(errors.WithStack(err))
	}

	fmt.Print(string(bytes))

}

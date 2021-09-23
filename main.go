package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

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

	if resp.StatusCode == 429 {
		os.Exit(1)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(errors.WithStack(err))
	}
	defer resp.Body.Close()
	var pData PurpleDate
	//fmt.Println(string(data))
	err = json.Unmarshal([]byte(data), &pData)
	if err != nil {
		panic(errors.WithStack(err))
	}

	//fmt.Println(pData.Fields)
	// [ID pm pm_cf_1 pm_atm age pm_0 pm_1 pm_2 pm_3 pm_4
	//  pm_5 pm_6 conf pm1 pm_10 p1 p2 p3 p4 p5
	// p6 Humidity Temperature Pressure Elevation Type Label Lat Lon Icon
	// isOwner Flags Voc Ozone1 Adc CH]
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
		//}

		//humidity, ok := row[21].(float64)
		//if !ok {
		//	panic(errors.New("Humidity did not Convert"))
		//}
		pm, ok := row[1].(float64)
		if !ok {
			panic(errors.New("PM did not Convert"))
		}
		temp, ok := row[22].(float64)
		if !ok {
			panic(errors.New("Temp did not Convert"))
		}
		temp = temp - 8 // on AVG the Temp is over by 8

		age, ok := row[4].(float64)
		if !ok {
			panic(errors.New("Age did not convert"))
		}
		if age > 60*60 {
			// fmt.Println("Skipping Too old AGE:", row[0], age, row[4])
			continue
		}

		//fmt.Printf("StationID:%v, PM2.5:%v Temp: %v, Humidity: %v AGE: %v\n", stationID, pm, temp, humidity, age)
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

	//(52°F − 32) × 5/9
	fer := sumTemp / count
	cel := (fer - 32) * 5 / 9
	o := output{
		Version:  "1",
		FullText: fmt.Sprintf("Air Quality: %0.1fpm Temp:%0.1f°F %0.1f°C   Nodes: %v  ", pmavg, fer, cel, count),
		Color:    color,
	}
	bytes, err := json.Marshal(o)
	if err != nil {
		panic(errors.WithStack(err))
	}

	fmt.Println(string(bytes))
}

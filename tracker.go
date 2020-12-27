package main

import (
	"encoding/json"
	"flag"
	//"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	//"strconv"
	//"encoding/json"
)

func initLogging() {
	profile := flag.String("profile", "test", "Environment profile")
	flag.Parse()

	textFormat := log.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}
	jsonFormat := log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyTime:  "@timestamp",
			log.FieldKeyLevel: "log.level",
			log.FieldKeyMsg:   "message",
		},
	}

	if *profile == "dev" {
		log.SetFormatter(&textFormat)
	} else {
		log.SetFormatter(&jsonFormat)
	}
}

func getFields(jsonMessage string) log.Fields {
	fields := log.Fields{}
	json.Unmarshal([]byte(jsonMessage), &fields)
	return fields
}

type Payload struct {
	Timestamp string `json:"@timestamp"`
	LogLevel  string `json:"log.level"`
	Products  []struct {
		ProductName string `json:"productName"`
		ProductID   string `json:"productId"`
		StoreInfo   struct {
			Distance    interface{} `json:"distance"`
			Latitude    float64     `json:"latitude"`
			Longitude   float64     `json:"longitude"`
			PhoneNumber struct {
				AreaCode             string `json:"AreaCode"`
				FormattedPhoneNumber string `json:"FormattedPhoneNumber"`
				LineNumber           string `json:"LineNumber"`
				Prefix               string `json:"Prefix"`
			} `json:"phoneNumber"`
			Quantity int `json:"quantity"`
			StoreID  int `json:"storeId"`
		} `json:"storeInfo"`
	} `json:"products"`
	URL string `json:"url"`
}

func main() {
	initLogging()

	client := &http.Client{}

	//req, err := http.NewRequest("GET", "https://www.abc.virginia.gov/webapi/inventory/mystore?storeNumbers=416&productCodes=018006", nil)
	req, err := http.NewRequest("GET", "https://www.abc.virginia.gov/webapi/inventory/mystore", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")
	q := req.URL.Query()
	q.Add("storeNumbers", "416")
	q.Add("productCodes", "018006")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	//fmt.Printf("%s", body)
	//fmt.Println(string(body))
	//log.Info(&body)
	//log.Field{"message": &body}

	//log.WithFields(log.Fields{
	//	"json": &body,
	//}).Info("message")

	log.WithFields(getFields(string(body))).Info()

	//log.WithFields(log.Fields{
	//	"message": string(body),
	//}).Info("message")
	//
	//res, err := http.Get("https://www.abc.virginia.gov/webapi/inventory/mystore?storeNumbers=416&productCodes=018006")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//body, err := ioutil.ReadAll(res.Body)
	//res.Body.Close()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("%s", body)

}

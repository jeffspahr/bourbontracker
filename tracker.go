package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	//log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)


type PayloadIn struct {
	Products  []struct {
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

type PayloadOut struct {
	Timestamp string `json:"@timestamp"`
	ProductName string `json:"bt.productName"`
	ProductID   string `json:"bt.productId"`
	Latitude    float64 `json:"geo.lat""`
	Longitude   float64  `json:"geo.lon"`
	//Latitude    float64     `json:"bt.latitude"`
	//Longitude   float64     `json:"bt.longitude"`
	Quantity int `json:"bt.quantity"`
	StoreID  int `json:"bt.storeId"`
	StoreURL string `json:"bt.storeurl"`
}

func main() {

	//Load store list from a file into an array
	var stores[]string

	file, err := os.Open("stores")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		stores=append(stores, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//Load products we care about from a file into a map
	fileProducts, err := os.Open("products.json")
	if err != nil {
		log.Fatal(err)
	}
	defer fileProducts.Close()
	var productsList map[string]string
	if err := json.NewDecoder(fileProducts).Decode(&productsList); err != nil {
		log.Fatal(err)
	}
	//Create string of comma delimited products to be used in the query string
	productListString := ""
	for key := range productsList {
		if productListString == ""{
			productListString = key
		}
		productListString += "," + key
	}

	for h := 0; h < len(stores); h++ {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://www.abc.virginia.gov/webapi/inventory/mystore", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Accept", "application/json")
		q := req.URL.Query()
		q.Add("storeNumbers", stores[h])
		q.Add("productCodes", productListString)
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

		pIn := PayloadIn{}
		err = json.Unmarshal(body, &pIn)
		if err != nil {
			log.Fatal(err)
		}

		pOut := PayloadOut{}

		for i := range pIn.Products {
			pOut.Timestamp = time.Now().Format(time.RFC3339)
			pOut.ProductName = productsList[pIn.Products[i].ProductID]
			pOut.ProductID = pIn.Products[i].ProductID
			pOut.Latitude = pIn.Products[i].StoreInfo.Latitude
			pOut.Longitude = pIn.Products[i].StoreInfo.Longitude
			pOut.Quantity = pIn.Products[i].StoreInfo.Quantity
			pOut.StoreID = pIn.Products[i].StoreInfo.StoreID
			pOut.StoreURL = "https://www.abc.virginia.gov/" + pIn.URL

			pOutJSON, err := json.Marshal(pOut)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s", pOutJSON)
		}

	}

}
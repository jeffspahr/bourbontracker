package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)


func main() {
	var stores[]string
	for i := 0; i < 500 ; i++ {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://www.abc.virginia.gov/webapi/inventory/mystore", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Accept", "application/json")
		q := req.URL.Query()
		q.Add("storeNumbers", strconv.Itoa(i))
		q.Add("productCodes", "018006")
		req.URL.RawQuery = q.Encode()
		println(req.URL.RawQuery)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		println(resp.StatusCode)
		if resp.StatusCode==200 {
			stores=append(stores, strconv.Itoa(i))
		}
	}
	//fmt.Println(stores)
	//fmt.Println(len(stores))
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("stores", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(stores); i++ {
		if _, err := f.WriteString(stores[i]+"\n"); err != nil {
			f.Close() // ignore error; Write error takes precedence
			log.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

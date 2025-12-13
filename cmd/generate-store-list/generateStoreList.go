package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	var stores []string

	// Add rate limiting to avoid API blocking
	baseDelay := 250 * time.Millisecond

	fmt.Println("Scanning for valid Virginia ABC store numbers (0-499)...")

	for i := 0; i < 500; i++ {
		// Sleep before each request except the first
		if i > 0 {
			time.Sleep(baseDelay)
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://www.abc.virginia.gov/webapi/inventory/mystore", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-type", "application/json")
		req.Header.Add("Accept", "application/json")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Add("Referer", "https://www.abc.virginia.gov/")
		q := req.URL.Query()
		q.Add("storeNumbers", strconv.Itoa(i))
		q.Add("productCodes", "018006") // Sample product code
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode == 200 {
			stores = append(stores, strconv.Itoa(i))
			fmt.Printf("✓ Store %d is valid (%d/%d)\n", i, len(stores), i+1)
		} else if i%50 == 0 {
			// Show progress every 50 stores
			fmt.Printf("  Scanned %d/%d stores, found %d valid...\n", i+1, 500, len(stores))
		}

		resp.Body.Close()
	}
	fmt.Printf("\n✅ Scan complete! Found %d valid stores out of 500\n", len(stores))
	fmt.Println("Writing stores to 'stores' file...")

	// If the file doesn't exist, create it, or append to the file
	file, err := os.OpenFile("stores", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(stores); i++ {
		if _, err := file.WriteString(stores[i] + "\n"); err != nil {
			file.Close() // ignore error; Write error takes precedence
			log.Fatal(err)
		}
	}
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ Successfully wrote %d store numbers to 'stores' file\n", len(stores))
}

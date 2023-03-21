package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type SteamResponse struct {
	Success     bool   `json:"success"`
	LowestPrice string `json:"lowest_price"`
	Volume      string `json:"volume"`
	MedianPrice string `json:"median_price"`
}

type ItemResponse struct {
	Success bool    `json:"success"`
	Price   float64 `json:"price"`
	Name    string  `json:"name"`
}

type Response struct {
	Items []ItemResponse
}

type ItemPayload struct {
	Name string
	Uri  string
}

type Payload struct {
	Items []ItemPayload
}

func steamHandler(w http.ResponseWriter, r *http.Request) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}

	reqBody, _ := ioutil.ReadAll(r.Body)
	var payload Payload
	json.Unmarshal(reqBody, &payload)

	var response []ItemResponse

	for i, item := range payload.Items {
		currentItem := getSteamItem(c, item.Uri, item.Name)

		response = append(response, currentItem)
		if i == len(payload.Items)-1 {
			break
		}
		time.Sleep(2 * time.Second)
	}

	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)

}

func normalizeGermanFloatString(old string) string {
	s := strings.Replace(old, ",", ".", -1)
	s = strings.Replace(s, "--", "00", -1)
	return strings.Replace(s, ".", "", 1)
}

func getSteamItem(client http.Client, itemUri string, itemName string) ItemResponse {
	var item ItemResponse

	resp, err := client.Get(itemUri)

	if err != nil {
		fmt.Printf("Error %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	var response SteamResponse
	json.Unmarshal([]byte(body), &response)

	lowestPriceFloat, err := strconv.ParseFloat(normalizeGermanFloatString(strings.TrimSuffix(response.LowestPrice, "€")), 32)
	item.Price = lowestPriceFloat / 100
	item.Name = itemName
	item.Success = response.Success

	return item

}

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	http.HandleFunc("/api/items", steamHandler)
	log.Printf("About to listen on %s. Go to https://127.0.0.1%s/", listenAddr, listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

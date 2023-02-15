package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Company struct {
	Id                 int                `json:"Id"`
	Name               string             `json:"Name"`
	Address            string             `json:"Address"`
	Zip                string             `json:"Zip"`
	City               string             `json:"City"`
	CountryId          int                `json:"CountryId"`
	Tel                string             `json:"Tel"`
	VAT                string             `json:"VAT"`
	ProductionEntities []ProductionEntity `json:"ProductionEntities"`
}

type ProductionEntity struct {
	Id      int    `json:"Id"`
	Name    string `json:"Name"`
	Address string `json:"Address"`
	Zip     string `json:"Zip"`
	City    string `json:"City"`
	Tel     string `json:"Tel"`
}

type Data struct {
	Id      int       `json:"Id"`
	Name    string    `json:"Name"`
	Company []Company `json:"Companies"`
}

func api(url string) []Data {

	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var data []Data

	err := json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	// for _, companyData := range data {
	// 	fmt.Println("companies: ", companyData.Company)
	// }

	return data
}

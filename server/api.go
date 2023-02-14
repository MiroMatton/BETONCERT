package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Companies struct {
	Id        int    `json:"Id"`
	Name      string `json:"Name"`
	Address   string `json:"Address"`
	Zip       string `json:"Zip"`
	City      string `json:"City"`
	CountryId int    `json:"CountryId"`
	Tel       string `json:"Tel"`
	VAT       string `json:"VAT"`
}

type Product struct {
	Id        int         `json:"Id"`
	Name      string      `json:"Name"`
	Companies []Companies `json:"Companies"`
}

func api(url string) Product {

	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var product []Product

	err := json.Unmarshal(body, &product)
	if err != nil {
		panic(err)
	}

	// for _, companyData := range product {
	// 	fmt.Println("company NAME: ", companyData.Name)
	// 	fmt.Println("company ID: ", companyData.Id)
	// }

	return product[0]

}

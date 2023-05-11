package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
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
	CategoryId         int                `json:"categoryId"`
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

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var data []Data

	e := json.Unmarshal(body, &data)
	if e != nil {
		panic(err)
	}

	// for _, companyData := range data {
	// 	fmt.Println("companies: ", companyData.Company)
	// }

	return data
}

type Entity struct {
	Id          int           `json:"Id"`
	Name        string        `json:"Name"`
	Certificate []Certificate `json:"Certificates"`
}

type Certificate struct {
	ID                       int     `bson:"ID"`
	Product                  string  `bson:"Product"`
	CertificateNumber        string  `bson:"CertificateNumber"`
	Standard                 string  `bson:"Standard"`
	TechnicalSpecification   *string `bson:"TechnicalSpecification"`
	CertificateReport        int     `bson:"CertificateReport"`
	SectorID                 int     `bson:"SectorId"`
	StatusID                 int     `bson:"StatusId"`
	NotLicensed              bool    `bson:"NotLicensed"`
	NotLicensedMessage       *string `bson:"NotLicensedMessage"`
	CertificationStatusID    int     `bson:"CertificationStatusId"`
	CertificationNotLicensed bool    `bson:"CertificationNotLicensed"`
	Suspended                bool    `bson:"Suspended"`
}

func certApi(companyId int, categoryId int) []Entity {

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.Printf("panic: %v\n%s", r, buf)
		}
	}()

	url := fmt.Sprintf("https://extranet.be-cert.be/api/HomePage/GetProductsTreeBranchForCompanyAndSector?languageIsoCode=en&treeFilters={%%22companyId%%22:%d,%%22sectorId%%22:%d,%%22certificationType%%22:%%22*%%22}", companyId, categoryId)

	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var certificate []Entity

	e := json.Unmarshal(body, &certificate)
	if e != nil {
		panic(err)
	}

	return certificate
}

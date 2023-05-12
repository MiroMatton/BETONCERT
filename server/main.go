package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	config := GetConfig()
	uri := fmt.Sprintf("mongodb+srv://%s:%s@alpha.mb8vfo3.mongodb.net/?retryWrites=true&w=majority", config.User, config.Password)
	//fmt.Println(uri)

	// connectie test
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("API running on : http://localhost:8080")
	server(client, ctx)
}

func company(client *mongo.Client, ctx context.Context) {
	url := "https://extranet.be-cert.be/api/HomePage/GetCertificateHoldersTree?languageIsoCode=en&treeFilters=%7B%22certificationType%22%3A%22*%22%7D"
	data := api(url)

	productsCollection := client.Database("demo").Collection("companiesTest")

	var err error // declare an err variable of type error

	for _, category := range data {
		for _, company := range category.Company {
			company.CategoryId = category.Id

			fmt.Println("Company:", company.Name, "was added")
			_, err = productsCollection.InsertOne(ctx, company)
			if err != nil {
				panic(err)
			}
		}
	}
}

func processProductionEntities(client *mongo.Client, ctx context.Context) {
	// Get the companies collection from the "demo" database
	companiesCollection := client.Database("demo").Collection("companiesTest")

	// Get the certificates collection from the "demo" database
	certificatesCollection := client.Database("demo").Collection("certificates")

	// Find all documents in the collection
	cursor, err := companiesCollection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	// Decode the documents into an array of Company structs
	var companies []Company
	if err = cursor.All(ctx, &companies); err != nil {
		panic(err)
	}

	for _, company := range companies {

		data := certApi(company.Id, company.CategoryId)

		if data == nil || len(data) == 0 {
			continue
		}

		for _, certificateCompany := range data {
			for _, certificate := range certificateCompany.Certificate {
				certificate.CompanyId = certificateCompany.Id

				fmt.Println("Certificate:", company.Name, "was added")
				_, err = certificatesCollection.InsertOne(ctx, certificate)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func getCertificates(client *mongo.Client, ctx context.Context) ([]bson.M, error) {
	// Access the "companiesTest" collection from the database
	catsCollection := client.Database("demo").Collection("certificates")
	cursor, err := catsCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	var cats []bson.M
	if err = cursor.All(ctx, &cats); err != nil {
		return nil, err
	}
	return cats, nil
}

func server(client *mongo.Client, ctx context.Context) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/certificates", func(w http.ResponseWriter, r *http.Request) {
		// Get data from MongoDB
		results, err := getCertificates(client, ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Encode data as JSON and write to response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", mux))
}

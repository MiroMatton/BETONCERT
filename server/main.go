package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func getCertificates(client *mongo.Client, ctx context.Context, page int, favouriteIDs []int, query string) ([]Certificate, int, error) {
	// Access the "certificates" collection from the database
	certCollection := client.Database("demo").Collection("certificates")

	// Calculate the number of documents to skip based on the page number
	perPage := 25
	skip := (page - 1) * perPage

	// Set up the options for the MongoDB query and filter by product if a query is provided
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(perPage))
	var filter bson.M
	if len(query) > 0 {
		filter = bson.M{"product": primitive.Regex{Pattern: query, Options: "i"}}
	}

	// Execute the query and retrieve the result set
	cursor, err := certCollection.Find(ctx, filter, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the result set to a slice of Certificate
	var certs []Certificate
	if err = cursor.All(ctx, &certs); err != nil {
		return nil, 0, err
	}

	// Check if each certificate is in the favourite list and set the isFavourite field
	for i, cert := range certs {
		for _, favID := range favouriteIDs {
			if cert.ID == favID {
				certs[i].IsFavourite = true
				break
			}
		}
	}

	// Calculate the total count of documents matching the query
	totalCount, err := certCollection.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	// Calculate total number of pages based on the total count and documents per page
	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

	return certs, totalPages, nil
}

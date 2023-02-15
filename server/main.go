package main

import (
	"context"
	"fmt"
	"log"

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

	url := "https://extranet.be-cert.be/api/HomePage/GetCertificateHoldersTree?languageIsoCode=en&treeFilters=%7B%22certificationType%22%3A%22*%22%7D"
	data := api(url)

	productsCollection := client.Database("demo").Collection("companies")

	for _, companyData := range data {
		for _, company := range companyData.Company {
			_, err = productsCollection.InsertOne(ctx, company)
			if err != nil {
				panic(err)
			}
			fmt.Println("company: ", company, "was added")
		}
	}

	// catsCollection := client.Database("demo").Collection("cats")
	// cursor, err := catsCollection.Find(ctx, bson.M{})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// var cats []bson.M
	// if err = cursor.All(ctx, &cats); err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(cats)
}

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

func certificatesList(client *mongo.Client, ctx context.Context, page int, perPage int, query string, favouriteIDs []int, mode string, activeCategories []int) ([]Certificate, int, error) {
	// Access the "certificates" collection from the database
	certCollection := client.Database("demo").Collection("certificates")

	// Calculate the number of documents to skip based on the page number
	skip := (page - 1) * perPage

	// Set up the options for the MongoDB query
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(perPage))

	// Declare variables for the filter and totalCount
	var filter bson.M
	var totalCount int64

	// Combine the search criteria depending on whether there are active categories or a search query
	if len(query) > 0 {
		if len(activeCategories) > 0 {
			filter = bson.M{"product": primitive.Regex{Pattern: query, Options: "i"}, "sectorid": bson.M{"$in": activeCategories}}
		} else {
			filter = bson.M{"product": primitive.Regex{Pattern: query, Options: "i"}}
		}
	} else if len(activeCategories) > 0 {
		filter = bson.M{"sectorid": bson.M{"$in": activeCategories}}
	}

	// Find certificates based on the mode
	var cursor *mongo.Cursor
	var err error
	switch mode {
	case "favorites":
		// Find all certificates with IDs in the "favourites" array
		cursor, err = certCollection.Find(ctx, bson.M{"id": bson.M{"$in": favouriteIDs}}, opts)
		if err != nil {
			return nil, 0, err
		}
		totalCount, err = certCollection.CountDocuments(ctx, bson.M{"id": bson.M{"$in": favouriteIDs}})
		if err != nil {
			return nil, 0, err
		}
	case "all":
		// Find all certificates that match the filter
		cursor, err = certCollection.Find(ctx, filter, opts)
		if err != nil {
			return nil, 0, err
		}
		totalCount, err = certCollection.CountDocuments(ctx, filter)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, fmt.Errorf("invalid mode '%s'", mode)
	}
	defer cursor.Close(ctx)

	// Convert the result set to a slice of Certificate and set the IsFavourite field to true for each one that's in the favourites list (if mode is "favourites")
	var certs []Certificate
	for cursor.Next(ctx) {
		var cert Certificate
		if err = cursor.Decode(&cert); err != nil {
			return nil, 0, err
		}
		for _, favID := range favouriteIDs {
			if cert.ID == favID {
				cert.IsFavourite = true
				break
			}
		}
		certs = append(certs, cert)
	}

	// Calculate total number of pages based on the total count and documents per page
	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))
	return certs, totalPages, nil
}

func getUserFavorites(client *mongo.Client, ctx context.Context, id string) ([]int, error) {
	collection := client.Database("demo").Collection("users")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectId}

	var result struct {
		FavoriteCertificates []int `bson:"favoriteCertificates"`
	}

	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.FavoriteCertificates, nil
}

func getCompanyByID(client *mongo.Client, ctx context.Context, id int) (*Company, error) {
	collection := client.Database("demo").Collection("companiesTest")

	filter := bson.M{"productionentities.id": id}

	result := collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var company Company
	err := result.Decode(&company)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

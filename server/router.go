package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
        if r.Method == "OPTIONS" {
            return
        }
        next.ServeHTTP(w, r)
    })
}

func server(client *mongo.Client, ctx context.Context) {
    r := mux.NewRouter()

	r.Use(corsMiddleware)

	r.HandleFunc("/api/certificates", getCertificatesHandler(client, ctx)).Methods("GET")
    r.HandleFunc("/api/certificates/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		toggleFavouriteHandler(w, r, client)
	}).Methods("PUT")	

    log.Fatal(http.ListenAndServe(":8080", r))
}

func getCertificatesHandler(client *mongo.Client, ctx context.Context) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract page and perPage parameters from query string
        pageStr := r.URL.Query().Get("page")
        favourites, _ := getUserFavorites(client, ctx, "645ff9c78f9b2d306a6d52ff")

        // Convert parameters to integers (default to 1 and 25 if not provided)
        page, _ := strconv.Atoi(pageStr)
        if page < 1 {
            page = 1
        }

        // Get data from MongoDB
        results, err := getCertificates(client, ctx, page, favourites)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Encode data as JSON and write to response
        if err := json.NewEncoder(w).Encode(results); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }
}

func toggleFavouriteHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) {
    // Extract certificate ID from URL path
    id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/certificates/"))
    if err != nil {
        http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
        return
    }

    // Extract isFavourite value from request body
    var reqBody struct {
        IsFavourite bool `json:"isFavourite"`
    }
    if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Update favorite status for user in MongoDB
    userCollection := client.Database("demo").Collection("users")
	objectID, err := primitive.ObjectIDFromHex("645ff9c78f9b2d306a6d52ff")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    filter := bson.M{"_id": objectID}
    update := bson.M{}
    if reqBody.IsFavourite {
        update = bson.M{"$addToSet": bson.M{"favoriteCertificates": id}}
    } else {
        update = bson.M{"$pull": bson.M{"favoriteCertificates": id}}
    }
    _, err = userCollection.UpdateOne(context.TODO(), filter, update)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "message": fmt.Sprintf("Favourite status for certificate %d updated for user %s", reqBody.IsFavourite, id),
    }); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
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
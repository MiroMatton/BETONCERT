package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func corsMiddleware(next http.Handler) http.Handler {
	headersOk := handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token", "X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	return handlers.CORS(headersOk, originsOk, methodsOk)(next)
}

func server(client *mongo.Client, ctx context.Context) {
	r := mux.NewRouter()

	r.HandleFunc("/api/certificates", certificatesHandler(client, ctx)).Methods("GET")
	r.HandleFunc("/api/favourite/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		toggleFavouriteHandler(w, r, client)
	}).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(r)))
}

func certificatesHandler(client *mongo.Client, ctx context.Context) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract query parameter from query string
		query := r.URL.Query().Get("q")

		// Set default values for pagination parameters
		page := 1
		perPage := 25

		// Parse page and per_page parameters from query string
		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				http.Error(w, "Invalid page number", http.StatusBadRequest)
				return
			}
		}

		if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
			var err error
			perPage, err = strconv.Atoi(perPageStr)
			if err != nil || perPage <= 0 {
				http.Error(w, "Invalid per_page number", http.StatusBadRequest)
				return
			}
		}

		// Get user favorites from MongoDB
		favourites, _ := getUserFavorites(client, ctx, "645ff9c78f9b2d306a6d52ff")

		// Retrieve data based on the query (if provided) or retrieve all data
		results, totalPages, err := getCertificates(client, ctx, page, favourites, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Encode data as JSON and write to response
		response := struct {
			Certificates []Certificate `json:"Certificates"`
			TotalPages   int           `json:"TotalPages"`
		}{
			Certificates: results,
			TotalPages:   totalPages,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func toggleFavouriteHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) {
	// Extract certificate ID from URL path
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/favourite/"))
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

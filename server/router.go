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
		// Extract query and page parameters from query string
		query := r.URL.Query().Get("q")
		pageStr := r.URL.Query().Get("page")
		fmt.Println(query)
		fmt.Println(pageStr)

		// Convert page parameter to integer (default to 1 if not provided)
		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}

		// Get user favorites from MongoDB
		favourites, _ := getUserFavorites(client, ctx, "645ff9c78f9b2d306a6d52ff")

		// Check whether there is a query parameter or not
		if len(query) > 0 {
			fmt.Println(query)
			// If there is a query parameter, retrieve data based on the query
			results, err := getCertificatesByProduct(client, ctx, page, favourites, query)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Encode data as JSON and write to response
			if err := json.NewEncoder(w).Encode(results); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// If there is no query parameter, retrieve all data
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

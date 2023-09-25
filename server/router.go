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
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	return handlers.CORS(headersOk, originsOk, methodsOk)(next)
}

func server(client *mongo.Client, ctx context.Context) {
	r := mux.NewRouter()

	r.HandleFunc("/api/certificates", certificatesHandler(client, ctx)).Methods("GET")
	r.HandleFunc("/api/favourite/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		toggleFavouriteHandler(w, r, client)
	}).Methods("PUT")
	r.HandleFunc("/api/company/{id:[0-9]+}", companyHandler(client, ctx)).Methods("GET")

	r.HandleFunc("/subscription", subscriptionsHandler(client, ctx)).Methods("POST")
	r.HandleFunc("/sendNotification", sendNotification(client, ctx)).Methods("GET")

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
		favourites, _ := getUserFavorites(client, ctx, "646f1bb8e6d409967e1a849c")

		// Get active categories from URI query string as a slice of ints
		activeCategoriesStr := r.URL.Query().Get("products")
		var activeCategories []int
		if activeCategoriesStr != "" {
			activeCategoriesStrArr := strings.Split(activeCategoriesStr, ",")
			for _, s := range activeCategoriesStrArr {
				i, err := strconv.Atoi(s)
				if err != nil {
					http.Error(w, "Invalid product ID", http.StatusBadRequest)
					return
				}
				activeCategories = append(activeCategories, i)
			}
		}

		var results []Certificate
		var totalPages int
		var err error

		// Check the value of the mode parameter and retrieve data accordingly
		var mode string = r.URL.Query().Get("mode")
		results, totalPages, err = certificatesList(client, ctx, page, perPage, query, favourites, mode, activeCategories)

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
	userCollection := client.Database("betonCert").Collection("users")
	objectID, err := primitive.ObjectIDFromHex("646f1bb8e6d409967e1a849c")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$addToSet": bson.M{"favoriteCertificates": id},
	}
	if !reqBody.IsFavourite {
		update = bson.M{
			"$pull": bson.M{"favoriteCertificates": id},
		}
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

func companyHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		company, err := getCompanyByID(client, ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(company)
	}
}

func subscriptionsHandler(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the subscription details from the request body
		var subscription map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Extract the notification key details from the subscription
		notificationKey := subscription

		// Save the subscription details to the user in the database
		userCollection := client.Database("betonCert").Collection("users")
		objectID, err := primitive.ObjectIDFromHex("646f1bb8e6d409967e1a849c")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filter := bson.M{"_id": objectID}
		update := bson.M{"$set": bson.M{"notificationKey": notificationKey}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Respond with a success message
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Subscription saved successfully",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func sendNotification(client *mongo.Client, ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCollection := client.Database("betonCert").Collection("users")
		objectID, err := primitive.ObjectIDFromHex("646f1bb8e6d409967e1a849c")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		filter := bson.M{"_id": objectID}
		var user map[string]interface{}
		if err := userCollection.FindOne(ctx, filter).Decode(&user); err != nil {
			if err == mongo.ErrNoDocuments {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		push(user)

		w.WriteHeader(http.StatusOK)
	}
}

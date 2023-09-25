package main

import (
	"encoding/json"
	"fmt"

	"github.com/SherClockHolmes/webpush-go"
)

func generateVAPIDKeys() (string, string, error) {
	privateKey, publicKey, _ := webpush.GenerateVAPIDKeys()
	return publicKey, privateKey, nil
}

func push(user map[string]interface{}) {
	config := GetConfig()

	subscriptionMap, ok := user["notificationKey"].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid notificationKey format")
		return
	}
	subscriptionBytes, err := json.Marshal(subscriptionMap)
	if err != nil {
		fmt.Println("Error marshaling subscription:", err)
		return
	}

	s := &webpush.Subscription{}
	json.Unmarshal(subscriptionBytes, s)

	// Send Notification
	resp, err := webpush.SendNotification([]byte("update certificate"), s, &webpush.Options{
		VAPIDPublicKey:  "BNOtWzRrDW8bwLSjwgbyvUwm5-aitqw0HJyL7Be-W6o_73Huy-KVqz4qNkBuoSQn71cHs9hBzCM8rj2GhdWL9CU",
		VAPIDPrivateKey: config.PrivateKey,
	})
	if err != nil {
		fmt.Println("Error notification didn't send:", err)
	} else {
		defer resp.Body.Close()
	}
}

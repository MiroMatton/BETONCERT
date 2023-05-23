package main

import (
	"encoding/json"
	"fmt"

	"github.com/SherClockHolmes/webpush-go"
)

func generateVAPIDKeys() (string, string, error) {
	publicKey := "BEskS8FtAmwXh88AOPD6T7JXYAyg_1ryvrflshNFOK9BAlqqm85OQ4xXA3FXnCGUOZ14glB0xZk1i6TThmJVVKE"
	privateKey := "itjiRwW-GfKA3VuB9y2xFdEsRdUKDSa5TT5yP8nKCf0"
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
	resp, err := webpush.SendNotification([]byte("timgay"), s, &webpush.Options{
		VAPIDPublicKey:  "BEskS8FtAmwXh88AOPD6T7JXYAyg_1ryvrflshNFOK9BAlqqm85OQ4xXA3FXnCGUOZ14glB0xZk1i6TThmJVVKE",
		VAPIDPrivateKey: config.PrivateKey,
	})
	if err != nil {

	}
	defer resp.Body.Close()
}

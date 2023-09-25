package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	gomail "gopkg.in/mail.v2"
)

func updateCertificatesCluster(client *mongo.Client, ctx context.Context) {
	// Get the companies collection from the "demo" database
	companiesCollection := client.Database("betonCert").Collection("companiesTest")

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

		data := getCertificates(company.Id, company.CategoryId)

		if data == nil || len(data) == 0 {
			continue
		}

		for _, certificateCompany := range data {
			updateCertificates(client, ctx, certificateCompany.Certificate)
		}
	}
}

func updateCertificates(client *mongo.Client, ctx context.Context, certificates []Certificate) {
	// Get the certificates collection from the "demo" database
	certificatesCollection := client.Database("betonCert").Collection("certificates")
	for _, certificate := range certificates {
		// Find the existing certificate with the same ID
		filter := bson.M{"id": certificate.ID}
		var existingCert Certificate
		err := certificatesCollection.FindOne(ctx, filter).Decode(&existingCert)
		if err != nil && err != mongo.ErrNoDocuments {
			panic(err)
		}

		// Check if the notlicensed, certificationnotlicensed, or suspended fields have changed
		notLicensedChanged := existingCert.NotLicensed != certificate.NotLicensed
		certNotLicensedChanged := existingCert.CertificationNotLicensed != certificate.CertificationNotLicensed
		suspendedChanged := existingCert.Suspended != certificate.Suspended

		// Update the certificate in the database
		update := bson.M{
			"$set": bson.M{
				"notlicensed":              certificate.NotLicensed,
				"certificationnotlicensed": certificate.CertificationNotLicensed,
				"suspended":                certificate.Suspended,
			},
		}
		upsert := true
		result, err := certificatesCollection.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
		if err != nil {
			panic(err)
		}

		// Check if the certificate was updated or inserted, and print a message if it was
		if result.MatchedCount > 0 || result.UpsertedCount > 0 {
			favourites, err := getUserFavorites(client, ctx, "645ff9c78f9b2d306a6d52ff")
			if err != nil {
				panic(err)
			}

			fmt.Println("Certificate:", certificate.ID, "was updated/added")
			if notLicensedChanged || certNotLicensedChanged || suspendedChanged {
				if notLicensedChanged {
					fmt.Println("Not licensed field changed from", existingCert.NotLicensed, "to", certificate.NotLicensed)
				}
				if certNotLicensedChanged {
					fmt.Println("Certification not licensed field changed from", existingCert.CertificationNotLicensed, "to", certificate.CertificationNotLicensed)
				}
				if suspendedChanged {
					fmt.Println("Suspended field changed from", existingCert.Suspended, "to", certificate.Suspended)
				}

				// Check if the updated certificate is in the user's favorites
				if contains(favourites, certificate.ID) {
					// Call the notification function
					notifyUser(certificate)
				}
			}
		} else {
			fmt.Println("Certificate:", certificate.ID, "was not updated/added")
		}

	}
}

func contains(array []int, id int) bool {
	for _, item := range array {
		if item == id {
			return true
		}
	}
	return false
}

func notifyUser(certificate Certificate) {
	fmt.Println("notify Thiery that the certficate: %s has been updated to valid %s", certificate.Product, certificate.NotLicensed)
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", "miromatton@gmail.com")

	// Set E-Mail receivers
	m.SetHeader("To", "to@example.com")

	// Set E-Mail subject
	m.SetHeader("Subject", "Gomail test subject")

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "This is Gomail test body")

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "from@gmail.com", "<email_password>")

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		panic(err)
	}

	return

}

func certficatesSeeder(client *mongo.Client, ctx context.Context) {
	// Get the companies collection from the "demo" database
	companiesCollection := client.Database("betonCert").Collection("companies")

	// Get the certificates collection from the "demo" database
	certificatesCollection := client.Database("betonCert").Collection("certificates")

	// Find all documents in the collection
	cursor, err := companiesCollection.Find(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	// Use a channel to synchronize the Goroutines
	ch := make(chan Certificate)

	for cursor.Next(ctx) {
		var company Company
		if err := cursor.Decode(&company); err != nil {
			fmt.Println(err)
		}

		go func(company Company) {
			data := getCertificates(company.Id, company.CategoryId)
			if data == nil || len(data) == 0 {
				return
			}

			for _, certificateCompany := range data {
				for _, certificate := range certificateCompany.Certificate {
					certificate.CompanyId = certificateCompany.Id

					fmt.Println("Certificate:", company.Name, "was added")
					if _, err = certificatesCollection.InsertOne(ctx, certificate); err != nil {
						fmt.Println(err)
					}

					// Send the certificate to the channel
					ch <- certificate
				}
			}
		}(company)
	}

	// Wait for all certificates to be inserted
	for range ch {
	}
}

func companySeeder(client *mongo.Client, ctx context.Context) {
	categories := getCompanies()

	companiesCollection := client.Database("betonCert").Collection("companies")

	ch := make(chan Company)

	for _, category := range categories {
		for _, company := range category.Company {
			company.CategoryId = category.Id

			go func(company Company) {
				fmt.Println("Company:", company.Name, "was added")
				if _, err := companiesCollection.InsertOne(ctx, company); err != nil {
					panic(err)
				}

				// Send the company to the channel
				ch <- company
			}(company)
		}
	}

	for range ch {
	}
}

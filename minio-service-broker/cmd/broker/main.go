package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/minio/minio-service-broker/auth"
)

func main() {
	req, err := http.NewRequest("PUT", "http://localhost:9001/one/two", nil)
	if err != nil {
		log.Fatal(err)
	}
	creds := auth.CredentialsV4{"minio", "minio123", "us-east-1"}
	creds.Sign(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.StatusCode)
}

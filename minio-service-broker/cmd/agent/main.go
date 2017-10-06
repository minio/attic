package main

import (
	"net/http"

	"github.com/minio/minio-service-broker/auth"
)

func main() {
	creds := auth.CredentialsV4{"minio", "minio123", "us-east-1"}
	http.ListenAndServe(":9001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !creds.IsSigned(r) {
			w.WriteHeader(http.StatusForbidden)
		}
		return
	}))
}

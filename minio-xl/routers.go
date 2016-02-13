/*
 * Minio Cloud Storage, (C) 2014 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"net/http"

	router "github.com/gorilla/mux"
	jsonrpc "github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"github.com/minio/minio-xl/pkg/xl"
)

// registerAPI - register all the object API handlers to their respective paths
func registerAPI(mux *router.Router, a API) {
	// root Router
	root := mux.NewRoute().PathPrefix("/").Subrouter()
	// Bucket router
	bucket := root.PathPrefix("/{bucket}").Subrouter()

	// Object operations
	bucket.Methods("HEAD").Path("/{object:.+}").HandlerFunc(a.HeadObjectHandler)
	bucket.Methods("PUT").Path("/{object:.+}").HandlerFunc(a.PutObjectPartHandler).Queries("partNumber", "{partNumber:[0-9]+}", "uploadId", "{uploadId:.*}")
	bucket.Methods("GET").Path("/{object:.+}").HandlerFunc(a.ListObjectPartsHandler).Queries("uploadId", "{uploadId:.*}")
	bucket.Methods("POST").Path("/{object:.+}").HandlerFunc(a.CompleteMultipartUploadHandler).Queries("uploadId", "{uploadId:.*}")
	bucket.Methods("POST").Path("/{object:.+}").HandlerFunc(a.NewMultipartUploadHandler).Queries("uploads", "")
	bucket.Methods("DELETE").Path("/{object:.+}").HandlerFunc(a.AbortMultipartUploadHandler).Queries("uploadId", "{uploadId:.*}")
	bucket.Methods("GET").Path("/{object:.+}").HandlerFunc(a.GetObjectHandler)
	bucket.Methods("PUT").Path("/{object:.+}").HandlerFunc(a.PutObjectHandler)
	// Not supported
	bucket.Methods("DELETE").Path("/{object:.+}").HandlerFunc(a.DeleteObjectHandler)

	// Bucket operations
	bucket.Methods("GET").HandlerFunc(a.GetBucketACLHandler).Queries("acl", "")
	bucket.Methods("GET").HandlerFunc(a.ListMultipartUploadsHandler).Queries("uploads", "")
	bucket.Methods("GET").HandlerFunc(a.ListObjectsHandler)
	bucket.Methods("PUT").HandlerFunc(a.PutBucketACLHandler).Queries("acl", "")
	bucket.Methods("PUT").HandlerFunc(a.PutBucketHandler)
	bucket.Methods("HEAD").HandlerFunc(a.HeadBucketHandler)
	bucket.Methods("POST").HandlerFunc(a.PostPolicyBucketHandler)
	// Not supported
	bucket.Methods("DELETE").HandlerFunc(a.DeleteBucketHandler)

	// Root operation
	root.Methods("GET").HandlerFunc(a.ListBucketsHandler)
}

// APIOperation container for individual operations read by Ticket Master
type APIOperation struct {
	ProceedCh chan struct{}
}

// API container for API and also carries OP (operation) channel
type API struct {
	OP        chan APIOperation
	XL        xl.Interface
	Anonymous bool // do not checking for incoming signatures, allow all requests
}

// getNewAPI instantiate a new minio API
func getNewAPI(anonymous bool) API {
	// ignore errors for now
	d, err := xl.New()
	fatalIf(err.Trace(), "Instantiating xl failed.", nil)

	return API{
		OP:        make(chan APIOperation),
		XL:        d,
		Anonymous: anonymous,
	}
}

func getAPIHandler(anonymous bool, api API) http.Handler {
	var mwHandlers = []MiddlewareHandler{
		TimeValidityHandler,
		IgnoreResourcesHandler,
		CorsHandler,
	}
	if !anonymous {
		mwHandlers = append(mwHandlers, SignatureHandler)
	}
	mux := router.NewRouter()
	registerAPI(mux, api)
	apiHandler := registerCustomMiddleware(mux, mwHandlers...)
	return apiHandler
}

func getServerRPCHandler(anonymous bool) http.Handler {
	var mwHandlers = []MiddlewareHandler{
		TimeValidityHandler,
	}
	if !anonymous {
		mwHandlers = append(mwHandlers, RPCSignatureHandler)
	}

	s := jsonrpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(serverRPCService), "Server")
	s.RegisterService(new(xlRPCService), "XL")
	mux := router.NewRouter()
	mux.Handle("/rpc", s)

	rpcHandler := registerCustomMiddleware(mux, mwHandlers...)
	return rpcHandler
}

// getControllerRPCHandler rpc handler for controller
func getControllerRPCHandler(anonymous bool) http.Handler {
	var mwHandlers = []MiddlewareHandler{
		TimeValidityHandler,
	}
	if !anonymous {
		mwHandlers = append(mwHandlers, RPCSignatureHandler)
	}

	s := jsonrpc.NewServer()
	codec := json.NewCodec()
	s.RegisterCodec(codec, "application/json")
	s.RegisterCodec(codec, "application/json; charset=UTF-8")
	s.RegisterService(new(controllerRPCService), "Controller")
	mux := router.NewRouter()
	// Add new RPC services here
	mux.Handle("/rpc", s)
	mux.Handle("/{file:.*}", http.FileServer(assetFS()))

	rpcHandler := registerCustomMiddleware(mux, mwHandlers...)
	return rpcHandler
}

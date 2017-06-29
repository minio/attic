/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"bytes"
	"encoding/json"
	"image"
	"net/http"
	"net/url"
	"sync"

	router "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	minio "github.com/minio/minio-go"
)

type xrayHandlers struct {
	sync.RWMutex

	// Object Storage handler.
	minioClient *minio.Client

	// Used for calculating motion detection.
	prevSR sensorRecord

	// Represents client response channel, sends client data.
	clntRespCh chan interface{}

	// Display memory channels.
	displayCh, displayRecvCh chan bool

	// Used for upgrading the incoming HTTP
	// wconnection into a websocket wconnection.
	upgrader websocket.Upgrader
}

var recorderMap = make(map[string]*motionRecorder)

func getRecordForClient(clientID string) *motionRecorder {
	if _, ok := recorderMap[clientID]; !ok {
		recorderMap[clientID] = &motionRecorder{}
	}
	mr, _ := recorderMap[clientID]
	return mr
}

// Detects face objects on incoming data.
func (v *xrayHandlers) detectObjects(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			errorIf(r.(error), "Recovered from a panic in detectObjects")
		}
	}()

	var fr frameRecord
	if err := json.Unmarshal(data, &fr); err != nil {
		errorIf(err, "Unable to unmarshal incoming frame record")
		v.clntRespCh <- XrayResult{
			Zoom: -1,
		}
		return
	}

	imgRect, frameID, err := fr.GetFullFrameRect()
	if err != nil {
		errorIf(err, "Unable to get image rect")
		v.clntRespCh <- XrayResult{
			Zoom: -1,
		}
		return
	}

	var motionDetected bool
	var optimalZoomFactor = -1
	if fr.Faces != nil {
		var faces []image.Rectangle
		faces, err = fr.GetFaceRectangles()
		if err != nil {
			errorIf(err, "Unable to get face rectangles")
			v.clntRespCh <- XrayResult{
				Zoom: -1,
			}
			return
		}

		// Get recorded frames.
		mr := getRecordForClient(fr.ClientID)
		mr.Append(&fr)

		// Check for motion detection.
		motionDetected = mr.DetectMotion()

		// Calculate optimal zoom factor for faces.
		optimalZoomFactor = calculateOptimalZoomFactor(faces, imgRect)

	} else if fr.Barcodes != nil {
		var barcodes []image.Rectangle
		barcodes, err = fr.GetBarcodeRectangles()
		if err != nil {
			errorIf(err, "Unable to get barcode rectangles")
			v.clntRespCh <- XrayResult{
				Zoom: -1,
			}
			return
		}

		// Motion is detected relevance is on for barcodes.
		motionDetected = len(barcodes) > 0

		// Calculate optimal zoom factor for barcodes.
		optimalZoomFactor = calculateOptimalZoomFactor(barcodes, imgRect)
	}

	pp := &url.URL{}
	if motionDetected {
		// Generate POST presigned URL.
		pp, err = v.newPresignedURL(genObjectName())
		if err != nil {
			errorIf(err, "Unable to generate presigned post policy")
			v.clntRespCh <- XrayResult{
				Zoom: -1,
			}
			return
		}
	}

	// Send the data to client.
	v.clntRespCh <- XrayResult{
		FrameID: frameID,
		Zoom:    optimalZoomFactor,
		URL:     pp.String(),
	}
}

// Detect detects metadata about the incoming data.
func (v *xrayHandlers) Detect(w http.ResponseWriter, r *http.Request) {
	wconn, err := v.upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorIf(err, "Unable to perform websocket upgrade the request.")
		return
	}

	wc := wConn{wconn}
	defer wc.Close()

	// Waiting on incoming reads.
	for {
		mt, data, err := wc.ReadMessage()
		if err != nil {
			errorIf(err, "Unable to read incoming message.")
			break
		}

		if mt == websocket.BinaryMessage {
			errorIf(err, "Invalid message type.")
			continue
		}

		// Ignore all other forms of incoming data.
		if bytes.Contains(data, []byte("sensorName")) {
			continue
		}

		go v.detectObjects(data)
		wc.WriteMessage(websocket.TextMessage, v.clntRespCh)
	}
}

// Initialize a new xray handlers.
func newXRayHandlers(clnt *minio.Client) *xrayHandlers {
	return &xrayHandlers{
		minioClient: clnt,
		clntRespCh:  make(chan interface{}, 15000),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}, // use default options
	}
}

// Configure xray handler.
func configureXrayHandler(mux *router.Router) http.Handler {
	// Register all xray handlers.
	registerXRayRouter(mux)

	// Register additional routers if any.
	return mux
}

// Register xray router.
func registerXRayRouter(mux *router.Router) {

	// Initialize minio client.
	clnt, err := newMinioClient()
	fatalIf(err, "Unable to initialize minio client")

	// Initialize xray handlers.
	xray := newXRayHandlers(clnt)

	// xray Router
	xrayRouter := mux.NewRoute().PathPrefix("/").Subrouter()

	// Currently there is only one handler.
	xrayRouter.Methods("GET").HandlerFunc(xray.Detect)
}

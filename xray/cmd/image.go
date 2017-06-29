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
	"image"
	"math"
	"strconv"
)

type frameStruct struct {
	ID        string `json:"id"`
	Format    string `json:"format"`
	Width     string `json:"width"`
	Height    string `json:"height"`
	Rotation  string `json:"rotation"`
	Timestamp string `json:"timestamp"`
}

type faceStruct struct {
	ID           string      `json:"id"`
	EulerY       string      `json:"eulerY"`
	EulerZ       string      `json:"eulerZ"`
	Height       string      `json:"height"`
	Width        string      `json:"width"`
	LeftEyeOpen  string      `json:"leftEyeOpen"`
	RightEyeOpen string      `json:"rightEyeOpen"`
	Smiling      string      `json:"similing"`
	FacePT1      pointStruct `json:"facePt1"`
	FacePT2      pointStruct `json:"facePt2"`
}

type barcodeStruct struct {
	ID         string      `json:"id"`
	BarcodePT1 pointStruct `json:"barcodePt1"`
	BarcodePT2 pointStruct `json:"barcodePt2"`
}

type pointStruct struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type frameRecord struct {
	ClientID string          `json:"client_uuid"`
	Frame    frameStruct     `json:"frame"`
	Faces    []faceStruct    `json:"faces"`
	Barcodes []barcodeStruct `json:"barcodes"`
}

// Extracts full frame rectangle from the incoming frame record.
func (fr *frameRecord) GetFullFrameRect() (image.Rectangle, int, error) {
	width, err := strconv.Atoi(fr.Frame.Width)
	if err != nil {
		return image.Rectangle{}, 0, err
	}

	height, err := strconv.Atoi(fr.Frame.Width)
	if err != nil {
		return image.Rectangle{}, 0, err
	}

	frameID, err := strconv.Atoi(fr.Frame.ID)
	if err != nil {
		return image.Rectangle{}, 0, err
	}

	return image.Rectangle{image.Point{}, image.Point{X: width, Y: height}}, frameID, nil
}

// Extracts all the barcode rectangles from the incoming frame record.
func (fr *frameRecord) GetBarcodeRectangles() ([]image.Rectangle, error) {
	var barcodes []image.Rectangle
	for _, barcode := range fr.Barcodes {
		x1, err := strconv.ParseFloat(barcode.BarcodePT1.X, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		y1, err := strconv.ParseFloat(barcode.BarcodePT1.Y, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		x2, err := strconv.ParseFloat(barcode.BarcodePT2.X, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		y2, err := strconv.ParseFloat(barcode.BarcodePT2.Y, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		barcodes = append(barcodes, image.Rectangle{
			image.Point{X: int(x1), Y: int(y1)}, image.Point{X: int(x2), Y: int(y2)},
		})
	}

	return barcodes, nil
}

// Extracts all the face rectangles from the incoming frame record.
func (fr *frameRecord) GetFaceRectangles() ([]image.Rectangle, error) {
	var faces []image.Rectangle
	for _, face := range fr.Faces {
		x1, err := strconv.ParseFloat(face.FacePT1.X, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		y1, err := strconv.ParseFloat(face.FacePT1.Y, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		x2, err := strconv.ParseFloat(face.FacePT2.X, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		y2, err := strconv.ParseFloat(face.FacePT2.Y, 64)
		if err != nil {
			return []image.Rectangle{}, err
		}
		faces = append(faces, image.Rectangle{
			image.Point{X: int(x1), Y: int(y1)}, image.Point{X: int(x2), Y: int(y2)},
		})
	}

	return faces, nil
}

// Rectangle represents custom rectangle implementation, provides
// additional methods for calculating threshold factors.
type Rectangle image.Rectangle

const zoomOutBorderSize = 50
const nozoomBorderSize = 75
const zoomBoost = 5

// Algorithm used here is pretty simple union of face rectangles is fitted
// into respectively smaller boxes, smallest box will return back the hightest
// zoom factor
func calculateOptimalZoomFactor(rects []image.Rectangle, boundingBox image.Rectangle) int {
	var final image.Rectangle
	for _, rect := range rects {
		final = final.Union(rect)
	}

	if final.Empty() {
		return -1 // Zoom out when nothing detected
	}

	nozoomBox := boundingBox.Inset(zoomOutBorderSize)
	zoomInBox1 := nozoomBox.Inset(nozoomBorderSize)

	inset := 0
	if zoomInBox1.Size().X < zoomInBox1.Size().Y {
		inset = zoomInBox1.Size().X / 4 / 2
	} else {
		inset = zoomInBox1.Size().Y / 4 / 2
	}

	zoomInBox2 := zoomInBox1.Inset(inset)
	zoomInBox3 := zoomInBox2.Inset(inset)

	if final.In(zoomInBox3) {
		return 3 * zoomBoost
	} else if final.In(zoomInBox2) {
		return 2 * zoomBoost
	} else if final.In(zoomInBox1) {
		return 1 * zoomBoost
	} else if final.In(nozoomBox) {
		return 0
	} else {
		return -1 * zoomBoost
	}
}

// Point represents - 2D points specified by its coordinates x and y.
type Point image.Point

// Radius - calculate the radius between the points.
func (p Point) Radius() float64 {
	return math.Sqrt(p.RadiusSq())
}

// RadiusSq - calculate raidus square X^2+Y^2
func (p Point) RadiusSq() float64 {
	return float64(p.X*p.X + p.Y*p.Y)
}

// Angle - calculate arc tangent of Y/X
func (p Point) Angle() float64 {
	return math.Atan2(float64(p.Y), float64(p.X))
}

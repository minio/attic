package cmd

import (
	"fmt"
	"image"
	"testing"
)

func TestZoomInSingle(t *testing.T) {

	boundingBox := image.Rect(0, 0, 960, 720)
	face := boundingBox.Inset(350)

	for {
		zoom := calculateOptimalZoomFactor([]image.Rectangle{face}, boundingBox)
		fmt.Println("For rectangle", face, "zoom =", zoom)
		if zoom == 0 {
			break
		}
		face = face.Inset(-10 * zoom) // grow box
	}
}

func TestZoomInTwoFaces(t *testing.T) {

	boundingBox := image.Rect(0, 0, 960, 720)
	face1 := boundingBox.Inset(350).Add(image.Point{X: 25})
	face2 := boundingBox.Inset(350).Sub(image.Point{X: 25})

	for {
		zoom := calculateOptimalZoomFactor([]image.Rectangle{face1, face2}, boundingBox)
		fmt.Println("For rectangles", face1, face2, "zoom =", zoom)
		if zoom == 0 {
			break
		}
		face1 = face1.Add(image.Point{X: 5 * zoom}) // shift box
	}
}

func TestZoomOutSingle(t *testing.T) {

	boundingBox := image.Rect(0, 0, 960, 720)
	face := boundingBox.Inset(10)

	for {
		zoom := calculateOptimalZoomFactor([]image.Rectangle{face}, boundingBox)
		fmt.Println("For rectangle", face, "zoom =", zoom)
		if zoom == 0 {
			break
		}
		face = face.Inset(-10 * zoom) // shrink box
	}
}

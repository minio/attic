package cmd

import (
	"encoding/json"
	"fmt"
	"image"
	"testing"
	"time"
)

var frame = 0
var timestamp = 99

const frameWidth = 960
const frameHeight = 720

func rectFromCenter(pt image.Point, size int) image.Rectangle {

	return image.Rect(pt.X-size, pt.Y-size, pt.X+size, pt.Y+size)
}

func getJSON(rects []image.Rectangle) string {

	frame++
	timestamp++

	jsonFaces := ""
	for id, rect := range rects {
		jsonFaces += fmt.Sprintf(`{ "id": "%d", "eulerY": "0.0", "eulerZ": "0.0", "width": "%d", "height": "%d", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "%d", "y": "%d"  }, "facePt2": {  "x": "%d",  "y": "%d" } }`,
			id+1, rect.Dx(), rect.Dy(), rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
		if id < len(rects)-1 {
			jsonFaces += ",\n             "
		}
	}

	return fmt.Sprintf(`{ "frame": { "id": "%d", "format": "17", "width": "%d", "height": "%d", "rotation": "2", "timestamp": "%d" }, "faces": [ %s ] }`, frame, frameWidth, frameHeight, timestamp, jsonFaces)
}

func TestMotion(t *testing.T) {

	mr := motionRecorder{}

	for size := 25; size < 20000/25; size++ {

		time.Sleep(time.Millisecond * 500)

		rect := rectFromCenter(image.Point{X: frameWidth/2 + size, Y: frameHeight / 2}, 25)
		jsontext := getJSON([]image.Rectangle{rect})

		var fr frameRecord
		if err := json.Unmarshal([]byte(jsontext), &fr); err != nil {
			panic("failed to unmarshal JSON")
		}

		mr.Append(&fr)
		if mr.DetectMotion() {
			fmt.Println("SNAPSHOT at threshold", mr.threshold())
		}
	}
}

func TestXorRects(t *testing.T) {

	expected, got := 0.0, 0.0
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(100, 100, 200, 200)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 2.0 * 100 * 100
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(300, 300, 400, 400)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 2.0 * 100 * 1
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(101, 100, 201, 200)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 2.0 * 100 * 50
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(150, 100, 250, 200)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 2.0*100*50 + 2.0*50*1
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(150, 101, 250, 201)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 2.0*102*1 + 2.0*99*1
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(99, 99, 201, 201)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 2.0*104*2 + 2.0*98*2
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(98, 98, 202, 202)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = (2.0*100*2 + 2.0*97*2) / 2
	got = sumAreas(XorRects(image.Rect(100, 100, 200, 200), image.Rect(101, 101, 199, 199)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}

	expected = 100*100 + 13
	got = sumAreas(XorRects(image.Rect(150, 150, 250, 250), image.Rect(128, 128, 273, 273)))
	if got != expected {
		t.Errorf("TestXorRects(): \nexpected %f\ngot      %f", expected, got)
	}
}

package cmd

import (
	"encoding/json"
	"fmt"
	"testing"
)

const jsontext = `{ "frame": { "id": "48", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "2295" }, "faces": [ { "id": "1", "eulerY": "0.0",
      "eulerZ": "16.868137", "height": "305.15756", "width": "244.12605", "leftEyeOpen": "-1.0", "rightEyeOpen": "0.588354", "similing": "0.007244766", "facePt1": { "x": "636.2853", "y": "332.01703" },
      "facePt2": { "x": "880.4114",  "y": "637.1746" } } ] }`

var jsonarray = []string{`{"frame":{"id":"88","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3412"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"3.6596656","height":"537.0032","width":"429.60254","leftEyeOpen":"0.6501604","rightEyeOpen":"0.7107066","smiling":"-1.0","facePt1":{"x":"320.31647","y":"612.7312"},"facePt2":{"x":"749.919","y":"1149.7344"}}]}`,
	`{"frame":{"id":"90","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3476"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.5272772","height":"519.4431","width":"415.55447","leftEyeOpen":"0.7430774","rightEyeOpen":"0.7009849","smiling":"-1.0","facePt1":{"x":"218.09044","y":"479.3399"},"facePt2":{"x":"633.6449","y":"998.783"}}]}`,
	`{"frame":{"id":"92","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3545"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.5272772","height":"518.5692","width":"414.85535","leftEyeOpen":"0.7430774","rightEyeOpen":"0.7009849","smiling":"-1.0","facePt1":{"x":"197.41748","y":"417.39032"},"facePt2":{"x":"612.2728","y":"935.95953"}}]}`,
	`{"frame":{"id":"93","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3588"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.5272772","height":"519.1202","width":"415.29614","leftEyeOpen":"0.7430774","rightEyeOpen":"0.7009849","smiling":"-1.0","facePt1":{"x":"194.8628","y":"391.93173"},"facePt2":{"x":"610.15894","y":"911.0519"}}]}`,
	`{"frame":{"id":"94","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3621"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.5272772","height":"519.5591","width":"415.64728","leftEyeOpen":"0.7430774","rightEyeOpen":"0.7009849","smiling":"-1.0","facePt1":{"x":"211.03943","y":"381.50964"},"facePt2":{"x":"626.6867","y":"901.0687"}}]}`,
	`{"frame":{"id":"95","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3656"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.5272772","height":"519.81586","width":"415.8527","leftEyeOpen":"0.7430774","rightEyeOpen":"0.7009849","smiling":"-1.0","facePt1":{"x":"220.2437","y":"379.87534"},"facePt2":{"x":"636.0964","y":"899.69116"}}]}`,
	`{"frame":{"id":"96","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3688"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.5272772","height":"519.92084","width":"415.93668","leftEyeOpen":"0.7430774","rightEyeOpen":"0.7009849","smiling":"-1.0","facePt1":{"x":"213.23903","y":"372.73636"},"facePt2":{"x":"629.1757","y":"892.6572"}}]}`,
	`{"frame":{"id":"97","format":"17","width":"1024","height":"768","rotation":"3","timestamp":"3724"},"faces":[{"id":"0","eulerY":"0.0","eulerZ":"1.3261887","height":"504.2792","width":"403.42337","leftEyeOpen":"0.5168264","rightEyeOpen":"0.98643833","smiling":"0.17594442","facePt1":{"x":"204.70055","y":"362.0135"},"facePt2":{"x":"608.1239","y":"866.29266"}}]}`}

var jsonCenterFace = []string{`{ "frame": { "id": "1", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "100" },
  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "50", "height": "50", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "455", "y": "335"  }, "facePt2": {  "x": "505",  "y": "385" } } ] }`,
	`{ "frame": { "id": "2", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "101" },
	  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "100", "height": "100", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "430", "y": "310"  }, "facePt2": {  "x": "530",  "y": "410" } } ] }`,
	`{ "frame": { "id": "3", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "102" },
	  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "150", "height": "150", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "405", "y": "285"  }, "facePt2": {  "x": "555",  "y": "435" } } ] }`,
	`{ "frame": { "id": "4", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "103" },
	  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "200", "height": "200", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "380", "y": "260"  }, "facePt2": {  "x": "580",  "y": "460" } } ] }`,
	`{ "frame": { "id": "5", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "104" },
	  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "250", "height": "250", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "355", "y": "235"  }, "facePt2": {  "x": "605",  "y": "485" } } ] }`}

var jsonZoomTwoFaces = []string{`{ "frame": { "id": "1", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "100" },
  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "50", "height": "50", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "355", "y": "335"  }, "facePt2": {  "x": "405",  "y": "385" } },
             { "id": "2", "eulerY": "0.0", "eulerZ": "0.0", "width": "50", "height": "50", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "555", "y": "335"  }, "facePt2": {  "x": "605",  "y": "385" } } ] }`,
	`{ "frame": { "id": "2", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "101" },
  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "100", "height": "100", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "330", "y": "310"  }, "facePt2": {  "x": "430",  "y": "410" } },
			 { "id": "2", "eulerY": "0.0", "eulerZ": "0.0", "width": "100", "height": "100", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "530", "y": "310"  }, "facePt2": {  "x": "630",  "y": "410" } } ] }`,
	`{ "frame": { "id": "3", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "102" },
  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "150", "height": "150", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "305", "y": "285"  }, "facePt2": {  "x": "455",  "y": "435" } },
			 { "id": "2", "eulerY": "0.0", "eulerZ": "0.0", "width": "150", "height": "150", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "505", "y": "285"  }, "facePt2": {  "x": "655",  "y": "435" } } ] }`,
	`{ "frame": { "id": "4", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "103" },
  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "200", "height": "200", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "280", "y": "260"  }, "facePt2": {  "x": "480",  "y": "460" } },
			 { "id": "2", "eulerY": "0.0", "eulerZ": "0.0", "width": "200", "height": "200", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "480", "y": "260"  }, "facePt2": {  "x": "680",  "y": "460" } } ] }`,
	`{ "frame": { "id": "5", "format": "17", "width": "960", "height": "720", "rotation": "2", "timestamp": "104" },
  "faces": [ { "id": "1", "eulerY": "0.0", "eulerZ": "0.0", "width": "250", "height": "250", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "255", "y": "235"  }, "facePt2": {  "x": "505",  "y": "485" } },
			 { "id": "2", "eulerY": "0.0", "eulerZ": "0.0", "width": "250", "height": "250", "leftEyeOpen": "-1.0", "rightEyeOpen": "-1.0", "similing": "0.0", "facePt1": { "x": "455", "y": "235"  }, "facePt2": {  "x": "705",  "y": "485" } } ] }`}

func TestParseJSON(t *testing.T) {

	testArray(t, []string{jsontext})

	testArray(t, jsonarray)

	testArray(t, jsonCenterFace)

	testArray(t, jsonZoomTwoFaces)
}

func testArray(t *testing.T, jsonarray []string) {

	for _, jsontext := range jsonarray {
		var fr frameRecord

		if err := json.Unmarshal([]byte(jsontext), &fr); err != nil {
			t.Errorf("TestZoomFactor(): failed to unmarshal JSON: %v", err)
		}

		boundingBox, _, err := fr.GetFullFrameRect()
		if err != nil {
			t.Errorf("TestZoomFactor() error: %v", err)
		}

		faces, err := fr.GetFaceRectangles()
		if err != nil {
			t.Errorf("TestZoomFactor() error: %v", err)
		}

		zoom := calculateOptimalZoomFactor(faces, boundingBox)
		fmt.Println("For Frame ID", fr.Frame.ID, " zoom =", zoom)
	}
}

/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Alice.
 *
 * Alice is free software: you can redistribute it and/or modify
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
 *
 */
package com.minio.io.alice;

import android.graphics.Rect;
import android.util.SparseArray;

import com.google.android.gms.vision.Frame;
import com.google.android.gms.vision.barcode.Barcode;
import com.google.android.gms.vision.face.Face;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;


/* Constructs a JSON object from the frame meta data and faces array for send to XRay
 *
 * Frame Metadata: https://developers.google.com/android/reference/com/google/android/gms/vision/Frame.Metadata
 * Face metadata: https://developers.google.com/android/reference/com/google/android/gms/vision/face/Face
 *
 */

public class FrameMetadata {

    // Represents Frame json object to be sent to the server.
    private JSONObject metaMap;

    public FrameMetadata(Frame.Metadata metadata, SparseArray<Face> facesArray, SparseArray<Barcode> barcodesArray) {
        if (facesArray == null) {
            try {
                metaMap = new JSONObject();
                metaMap.put("client_uuid",MainActivity.CLIENT_UUID);
                metaMap.put("frame", getMetaDataJSON(metadata));
                metaMap.put("barcodes", getBarcodesJSON(barcodesArray));
            } catch (JSONException e) {
                e.printStackTrace();
            }
        } else {
            try {
                metaMap = new JSONObject();
                metaMap.put("frame", getMetaDataJSON(metadata));
                metaMap.put("faces", getFacesJSON(facesArray));

            } catch (JSONException e) {
                e.printStackTrace();
            }
        }
    }

    public FrameMetadata(Frame.Metadata metadata, SparseArray<Face> facesArray) {
        new FrameMetadata(metadata, facesArray, null);
    }

    // Constructs frame metadata json object.
    private JSONObject getMetaDataJSON(Frame.Metadata metadata) {

        JSONObject meta = null;
        try {
            meta = new JSONObject();
            meta.put("id", Integer.toString(metadata.getId()));
            meta.put("format", Integer.toString(metadata.getFormat()));
            meta.put("width", Integer.toString(metadata.getWidth()));
            meta.put("height", Integer.toString(metadata.getHeight()));
            meta.put("rotation", Integer.toString(metadata.getRotation()));
            meta.put("timestamp", Long.toString(metadata.getTimestampMillis()));

        } catch (JSONException e) {
            e.printStackTrace();
        }
        return meta;
    }

    // Constructs list of barcodes json array object.
    private JSONArray getBarcodesJSON(SparseArray<Barcode> barcodes) {
        JSONArray barcodeArray = null;
        try {
            barcodeArray = new JSONArray();
            for (int i = 0; i < barcodes.size(); i++) {
                int key = barcodes.keyAt(i);
                Barcode barcode = barcodes.get(key);

                JSONObject bObject = new JSONObject();
                bObject.put("id", "1");

                Rect boundingBox = barcode.getBoundingBox();
                JSONObject pt1 = new JSONObject();
                JSONObject pt2 = new JSONObject();
                pt1.put("x", Float.toString(boundingBox.left));
                pt1.put("y", Float.toString(boundingBox.top));

                pt2.put("x", Float.toString(boundingBox.right));
                pt2.put("y", Float.toString(boundingBox.bottom));
                bObject.put("barcodePt1", pt1);
                bObject.put("barcodePt2", pt2);

                barcodeArray.put(i, bObject);
            }
        } catch (JSONException e) {
            e.printStackTrace();
        }

        return barcodeArray;
    }

    // Constructs list of faces json array object.
    private JSONArray getFacesJSON(SparseArray<Face> faces) {

        JSONArray faceArray = null;
        try {
            faceArray = new JSONArray();

            float xOffset, yOffset;
            float left, right, top, bottom;
            for (int i = 0; i < faces.size(); i++) {
                int key = faces.keyAt(i);
                Face face = faces.get(key);

                JSONObject fObject = new JSONObject();
                fObject.put("id", Integer.toString(face.getId()));
                fObject.put("eulerY", Float.toString(face.getEulerY()));
                fObject.put("eulerZ", Float.toString(face.getEulerZ()));
                fObject.put("height", Float.toString(face.getHeight()));
                fObject.put("width", Float.toString(face.getWidth()));

                fObject.put("leftEyeOpen", Float.toString(face.getIsLeftEyeOpenProbability()));
                fObject.put("rightEyeOpen", Float.toString(face.getIsRightEyeOpenProbability()));
                fObject.put("smiling", Float.toString(face.getIsSmilingProbability()));


                // Following block constructs the rectangular points of the face.
                float x = face.getPosition().x + face.getWidth() / 2.0f;
                float y = face.getPosition().y + face.getHeight() / 2.0f;

                xOffset = face.getWidth() / 2.0f;
                yOffset = face.getHeight() / 2.0f;

                left = x - xOffset;
                top = y - yOffset;
                right = x + xOffset;
                bottom = y + yOffset;

                JSONObject pt1 = new JSONObject();
                JSONObject pt2 = new JSONObject();
                pt1.put("x", Float.toString(left));
                pt1.put("y", Float.toString(top));

                pt2.put("x", Float.toString(right));
                pt2.put("y", Float.toString(bottom));
                fObject.put("facePt1", pt1);
                fObject.put("facePt2", pt2);

                faceArray.put(i, fObject);
            }

        } catch (JSONException e) {
            e.printStackTrace();
        }

        return faceArray;
    }

    public String toString() {
        return metaMap.toString();
    }

}

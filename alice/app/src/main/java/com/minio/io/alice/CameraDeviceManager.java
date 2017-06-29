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

import android.content.Context;
import android.util.Log;

import com.google.android.gms.vision.Detector;
import com.google.android.gms.vision.MultiDetector;
import com.google.android.gms.vision.MultiProcessor;
import com.google.android.gms.vision.barcode.Barcode;
import com.google.android.gms.vision.barcode.BarcodeDetector;
import com.google.android.gms.vision.face.Face;
import com.google.android.gms.vision.face.FaceDetector;
import com.google.android.gms.vision.text.TextRecognizer;

import java.io.IOException;

import static android.content.ContentValues.TAG;
import static android.hardware.Camera.Parameters.FOCUS_MODE_CONTINUOUS_VIDEO;

/*
    CameraDeviceManager manages the Camera device.
 */

public class CameraDeviceManager {

    private static int mCameraId;
    private Context context;
    private MainActivity mainActivity;
    private CameraSource mCameraSource;
    private static final int RC_HANDLE_GMS = 9001;
    private MultiProcessor.Factory<Face> graphicFaceTrackerFactory;
    private MultiProcessor.Factory<Barcode> barcodeTrackerFactory;
    private OcrDetectorProcessor ocrProcessor;

    PreviewCallback callback;
    FrameHandler mframeHandler;
    private boolean isCameraIdSet;

    public CameraDeviceManager(Context context, OcrDetectorProcessor ocrProcessor,
                               MultiProcessor.Factory<Face> faceTrackerFactory,
                               MultiProcessor.Factory<Barcode> barcodeTrackerFactory,
                               FrameHandler handler) {
        this.context = context;
        this.mainActivity = (MainActivity) context;
        this.ocrProcessor = ocrProcessor;
        this.graphicFaceTrackerFactory = faceTrackerFactory;
        this.barcodeTrackerFactory = barcodeTrackerFactory;
        this.mframeHandler = handler;
        callback = (MainActivity)context;
    }

    public static int getFacingCamera() {
        return mCameraId;
    }

    public void setFacingCamera(int cameraId) {
        this.mCameraId = cameraId;
    }

    /**
     * Creates and starts the camera.  Note that this uses a higher resolution in comparison
     * to other detection examples to enable the barcode detector to detect small barcodes
     * at long distances.
     */
    public void createCameraSource(int cameraId) {

        this.mCameraId = cameraId;

        BarcodeDetector barcodeDetector = new BarcodeDetector.Builder(context).build();
        // Custom barcode detector to grab metadata.
        Detector<Barcode> customBarcodeDetector = new CustomBarcodeDetector(context, barcodeDetector);
        customBarcodeDetector.setProcessor(
                new MultiProcessor.Builder<>(this.barcodeTrackerFactory).build());

        if (!barcodeDetector.isOperational()) {
            // Note: The first time that an app using the barcode or face API is installed on a
            // device, GMS will download a native libraries to the device in order to do detection.
            // Usually this completes before the app is run for the first time.  But if that
            // download has not yet completed, then the above call will not detect any barcodes
            // and/or faces.
            //
            // isOperational() can be used to check if the required native libraries are currently
            // available.  The detectors will automatically become operational once the library
            // downloads complete on device.
            Log.w(TAG, "Detector dependencies are not yet available.");
        }

        // A text recognizer is created to find text.  An associated processor instance
        // is set to receive the text recognition results and display graphics for each text block
        // on screen.
        TextRecognizer textRecognizer = new TextRecognizer.Builder(context).build();
        textRecognizer.setProcessor(this.ocrProcessor);

        if (!textRecognizer.isOperational()) {
            // Note: The first time that an app using a Vision API is installed on a
            // device, GMS will download a native libraries to the device in order to do detection.
            // Usually this completes before the app is run for the first time.  But if that
            // download has not yet completed, then the above call will not detect any text,
            // barcodes, or faces.
            //
            // isOperational() can be used to check if the required native libraries are currently
            // available.  The detectors will automatically become operational once the library
            // downloads complete on device.
            Log.w(TAG, "Text detector dependencies are not yet available.");

        }

        FaceDetector detector = new FaceDetector.Builder(context)
                .setClassificationType(FaceDetector.ALL_CLASSIFICATIONS)
                .build();

        // Custom face detector to grab metadata.
        Detector<Face> customFaceDetector = new CustomFaceDetector(context, detector);
        customFaceDetector.setProcessor(
                new MultiProcessor.Builder<>(this.graphicFaceTrackerFactory)
                        .build());

        if (!customFaceDetector.isOperational()) {
            // Note: The first time that an app using face API is installed on a device, GMS will
            // download a native library to the device in order to do detection.  Usually this
            // completes before the app is run for the first time.  But if that download has not yet
            // completed, then the above call will not detect any faces.
            //
            // isOperational() can be used to check if the required native library is currently
            // available.  The detector will automatically become operational once the library
            // download completes on device.
            Log.w(TAG, "Face detector dependencies are not yet available.");
        }

        MultiDetector multiDetector = new MultiDetector.Builder()
                .add(customBarcodeDetector)
                .add(textRecognizer)
                .add(customFaceDetector)
                .build();

        mCameraSource = new CameraSource.Builder(context, multiDetector)
                .setFacing(cameraId)
                .setRequestedFps(1.0f)
                .setFrameHandler(mframeHandler)
                .setFocusMode(FOCUS_MODE_CONTINUOUS_VIDEO)
                .build();
    }

    /**
     * Starts or restarts the camera source, if it exists.  If the camera source doesn't exist yet
     * (e.g., because onResume was called before the camera source was created), this will be called
     * again when the camera source is created.
     */
    public void startCameraSource() {

//        // check that the device has play services available.
//        int code = GoogleApiAvailability.getInstance().isGooglePlayServicesAvailable(
//                context);
//        if (code != ConnectionResult.SUCCESS) {
//            Dialog dlg =
//                    GoogleApiAvailability.getInstance().getErrorDialog(context, code, RC_HANDLE_GMS);
//            dlg.show();
//        }

        if (mCameraSource != null) {
            try {
                callback.startPreview(mCameraSource);
            } catch (IOException e) {
                Log.e(TAG, "Unable to start camera source.", e);
                mCameraSource.release();
                mCameraSource = null;
            }
        }
    }

    public void releaseCameraResource() {
        if (mCameraSource != null) {
            mCameraSource.release();
        }
    }
    public void pauseCameraResource() {
        if (mCameraSource != null) {
            mCameraSource.stop();
        }
    }
    public void swapCameraSource() {
        int switchCameraId =  (mCameraId == CameraSource.CAMERA_FACING_FRONT) ? CameraSource.CAMERA_FACING_BACK : CameraSource.CAMERA_FACING_FRONT;
        mCameraSource.release();
        mCameraId = switchCameraId;
        createCameraSource(switchCameraId);
        startCameraSource();
    }

    private void releaseCamera() {
        if (mCameraSource != null)
            mCameraSource.release();
        mCameraSource = null;
    }
    public void onPause() {
        releaseCamera();
    }

    public void onResume() {
        if (mCameraSource == null) {
            this.createCameraSource(mCameraId);
            this.startCameraSource();
        }
    }

    public void onDestroy() {
        releaseCamera();
    }
}

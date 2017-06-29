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
import android.graphics.Bitmap;
import android.graphics.Rect;
import android.os.SystemClock;
import android.util.SparseArray;

import com.google.android.gms.vision.Detector;
import com.google.android.gms.vision.Frame;
import com.google.android.gms.vision.barcode.Barcode;

/**
 * This is a wrapper around FaceDetector that captures meta data about frame and detected faces.
 */
public class CustomBarcodeDetector extends Detector<Barcode> {

    private Detector<Barcode> mDelegate;
    private AliceTask vTask;
    private MainActivity mainActivity;

    /**
     * Creates a custom face detector to wrap around underlying face detector
     */
    public CustomBarcodeDetector(Context context, Detector<Barcode> delegate) {
        mainActivity = (MainActivity) context;
        mDelegate = delegate;
    }

    @Override
    public void release() {
        mDelegate.release();
    }

    /**
     * Captures meta data about barcodes detected and affiliated frame info for the XRay server
     */
    @Override
    public SparseArray<Barcode> detect(Frame frame) {
        SparseArray<Barcode> barcodes = null;
        try {
            barcodes = mDelegate.detect(frame);
            if (barcodes.size() > 0) {
                MainActivity.isAliceAwake = true;
                MainActivity.prevFaceDetectionAt = SystemClock.elapsedRealtime();

                vTask = new AliceTask(frame.getMetadata(), null, barcodes.clone());
                vTask.execute();

                // Fetch the current frame rotation.
                int rotation = frame.getMetadata().getRotation();

                mainActivity.getPreview().increaseZoom(calculateOptimalZoomFactor(
                        frame.getMetadata(),
                        barcodes.clone(),
                        rotation));

                // Fetch the underlying frame bitmap.
                Bitmap bitmap = frame.getBitmap();

                // Rotate with proper orientation.
                bitmap = Utils.rotateBitmap(bitmap, rotation);

                // Save the face detected frames.
                FrameHandler mFrameHandler = mainActivity.getFrameHandler();
                mFrameHandler.addFramesToQueue(mFrameHandler.yuv2JPEG(bitmap));
            }
        } catch (Exception e) {
            throw e;
        }

        return barcodes;

    }

    @Override
    public boolean isOperational() {
        return mDelegate.isOperational();
    }

    @Override
    public boolean setFocus(int id) {
        return mDelegate.setFocus(id);
    }

    // Calculate optimal zoom factor depending on input factors
    // such as Frame metadata and list of faces.
    // This optimal zoom allows to control the zoom factor such that we
    // we do not overflow our face bounding boxes.
    private int calculateOptimalZoomFactor(Frame.Metadata metaData, SparseArray<Barcode> barcodes, int rotation) {
        Rect combinedRectangle = new Rect();
        final int zoomOutBorderSize = 50;
        final int nozoomBorderSize = 75;

        Rect boundingBox;
        if (rotation == 1 || rotation == 3)
            boundingBox = new Rect(0, 0, metaData.getHeight(), metaData.getWidth());
        else
            boundingBox = new Rect(0, 0, metaData.getWidth(), metaData.getHeight());

        try {
            for (int i = 0; i < barcodes.size(); i++) {
                int key = barcodes.keyAt(i);
                Barcode barcode = barcodes.get(key);
                combinedRectangle.union(barcode.getBoundingBox());
            }
        } catch (Exception e) {
            e.printStackTrace();
        }

        if (combinedRectangle.isEmpty()) {
            return -1;
        }

        Rect nozoomBox = new Rect(boundingBox);
        nozoomBox.inset(zoomOutBorderSize, zoomOutBorderSize);

        Rect zoomBox1 = new Rect(nozoomBox);
        zoomBox1.inset(nozoomBorderSize, nozoomBorderSize);

        int inset = 0;
        if (zoomBox1.width() < zoomBox1.height()) {
            inset = zoomBox1.width() / 8;
        } else {
            inset = zoomBox1.height() / 8;
        }

        Rect zoomBox2 = new Rect(zoomBox1);
        zoomBox2.inset(inset, inset);

        Rect zoomBox3 = new Rect(zoomBox2);
        zoomBox3.inset(inset, inset);

        if (zoomBox3.contains(combinedRectangle)) {
            return 3;
        } else if (zoomBox2.contains(combinedRectangle)) {
            return 2;
        } else if (zoomBox1.contains(combinedRectangle)) {
            return 1;
        } else if (nozoomBox.contains(combinedRectangle)) {
            // No zoom if we have a reached the maximum possible
            // bounding rectangle.
            return 0;
        }

        // With no zoom factor, we zoom out.
        return -1;
    }

}

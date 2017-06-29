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
import com.google.android.gms.vision.face.Face;

/**
 * This is a wrapper around FaceDetector that captures meta data about frame and detected faces.
 */
public class CustomFaceDetector extends Detector<Face> {

    private Detector<Face> mDelegate;
    private AliceTask vTask;
    private MainActivity mainActivity;

    /**
     * Creates a custom face detector to wrap around underlying face detector
     */
    public CustomFaceDetector(Context context, Detector<Face> delegate) {
        mainActivity = (MainActivity) context;
        mDelegate = delegate;
    }

    @Override
    public void release() {
        mDelegate.release();
    }

    /**
     * Captures meta data about faces detected and affiliated frame info for the XRay server
     */
    @Override
    public SparseArray<Face> detect(Frame frame) {
        SparseArray<Face> faces = null;
        try {
            faces = mDelegate.detect(frame);
            if (faces.size() > 0) {
                MainActivity.isAliceAwake = true;
                MainActivity.prevFaceDetectionAt = SystemClock.elapsedRealtime();

                // Fetch the current frame rotation.
                int rotation = frame.getMetadata().getRotation();

                mainActivity.getPreview().increaseZoom(calculateOptimalZoomFactor(frame.getMetadata(),
                        faces.clone(), rotation));

                // Fetch the underlying frame bitmap.
                Bitmap bitmap = frame.getBitmap();

                // Rotate with proper orientation.
                bitmap = Utils.rotateBitmap(bitmap, rotation);

                // Save the face detected frames.
                FrameHandler mFrameHandler = mainActivity.getFrameHandler();
                mFrameHandler.addFramesToQueue(mFrameHandler.yuv2JPEG(bitmap));
            } else {
                mainActivity.getPreview().resetZoom();
            }
        } catch (Exception e) {
            throw e;
        }

        return faces;

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
    private int calculateOptimalZoomFactor(Frame.Metadata metaData, SparseArray<Face> faces, int rotation) {

        Rect combinedRectangle = new Rect();
        final int zoomOutBorderSize = 50;
        final int nozoomBorderSize = 75;

        Rect boundingBox;

        if (rotation == 1 || rotation == 3)
            boundingBox = new Rect(0, 0, metaData.getHeight(), metaData.getWidth());
        else
            boundingBox = new Rect(0, 0, metaData.getWidth(), metaData.getHeight());

        try {
            float xOffset, yOffset;
            float left, right, top, bottom;
            for (int i = 0; i < faces.size(); i++) {
                int key = faces.keyAt(i);
                Face face = faces.get(key);


                // Following block constructs the rectangular points of the face.
                float x = face.getPosition().x + face.getWidth() / 2.0f;
                float y = face.getPosition().y + face.getHeight() / 2.0f;

                xOffset = face.getWidth() / 2.0f;
                yOffset = face.getHeight() / 2.0f;

                left = x - xOffset;
                top = y - yOffset;
                right = x + xOffset;
                bottom = y + yOffset;

                Rect rect = new Rect(Math.round(left), Math.round(top), Math.round(right), Math.round(bottom));
                combinedRectangle.union(rect);
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

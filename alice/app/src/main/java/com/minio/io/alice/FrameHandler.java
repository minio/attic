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


import android.graphics.Bitmap;
import java.io.ByteArrayOutputStream;
import java.util.LinkedList;


/*
 This  handler's only job is to get frames from the CameraSource and store in a queue.
 As the queue is polled,frames are handed to server handler for passing to the
 XRay server
  */
public class FrameHandler implements Runnable {
    private LinkedList<byte[]> mQueue;
    private static final int MAX_BUFFER = 100;

    public FrameHandler() {
        mQueue = new LinkedList<>();
    }

    // Enqueue frames received from camera.
    public void addFramesToQueue(byte[] data) {
        synchronized (mQueue) {
            if (mQueue.size() == MAX_BUFFER) {
                mQueue.poll();
            }
            mQueue.add(data);
        }
    }

    // Converts Android's NV21 image format to RGBA_8888, and then
    // to the compressed JPEG format recognized by the server
    public byte[] yuv2JPEG(Bitmap bmp) {
        // Convert RGBA_8888 to ARGB_8888 compressed JPEG format
        ByteArrayOutputStream byteStream = new ByteArrayOutputStream();
        bmp.compress(Bitmap.CompressFormat.JPEG, 100, byteStream);
        bmp.recycle();
        return byteStream.toByteArray();
    }

    // Poll queue for new frames
    public byte[] getFrame() {
        synchronized (mQueue) {
            return mQueue.poll();
        }
    }

    // Send incoming frames to serverHandler.
    @Override
    public void run() {}
}

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

import android.os.AsyncTask;
import android.util.Log;
import android.util.SparseArray;

import com.google.android.gms.vision.Frame;
import com.google.android.gms.vision.barcode.Barcode;
import com.google.android.gms.vision.face.Face;

/**
 * AsyncTask that sends the video buffer over to the Xray Server.
 */

public class AliceTask extends AsyncTask<Void, Integer, String> {
    byte[] binaryData;
    String textData;
    boolean textMessage;

    public AliceTask(byte[] data) {
        this.binaryData = data;
    }

    public AliceTask(String data) {
        this.textMessage = true;
        this.textData = data;
    }

    public AliceTask(Frame.Metadata metadata, SparseArray<Face> faces, SparseArray<Barcode> barcodes) {
        this.textMessage = true;
        this.textData = new FrameMetadata(metadata, faces, barcodes).toString();
    }

    @Override
    protected void onPreExecute() {}

    @Override
    protected String doInBackground(Void ... params) {
        if(MainActivity.webSocket == null) {
            if (XDebug.LOG) {
                Log.d(MainActivity.TAG, "Socket not connected");
            }
            return String.valueOf(R.string.COMPLETE);
        }

        if (textMessage)
            MainActivity.webSocket.sendPayload(textData);
        else
            MainActivity.webSocket.sendPayload(binaryData);

        return String.valueOf(R.string.COMPLETE);

    }

    @Override
    protected void onCancelled() {}

    protected void onPostExecute(String finish) {}
}

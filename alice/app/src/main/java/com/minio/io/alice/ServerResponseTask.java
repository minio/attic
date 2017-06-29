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
import android.os.AsyncTask;

/**
 * AsyncTask that responds to server commands to zoom in.
 */

public class ServerResponseTask extends AsyncTask<Void, Void, Void> {
    XrayResult serverResult;
    boolean setZoom = false;
    CameraSourcePreview mPreview;
    GraphicOverlay mServerOverlay;
    MainActivity mainActivity;

   ServerResponseTask(Context context, XrayResult xresult, GraphicOverlay serverOverlay, CameraSourcePreview preview) {
       mainActivity = (MainActivity) context;
       serverResult = xresult;
       mServerOverlay = serverOverlay;
       mPreview =  preview;
   }


    @Override
    protected void onPreExecute() {}

    @Override
    protected Void doInBackground(Void ... params) {
        if (serverResult != null) {
            if (serverResult.isReply() == true) {
                // Fetch the frames from the frame handler.
                FrameHandler mframeHandler = mainActivity.getFrameHandler();
                if (mframeHandler != null) {
                    // Image data can be null when getFrame() is
                    // initiated when there are no frames on the
                    // linked list.
                    byte[] imageData = mframeHandler.getFrame();
                    if (imageData != null) {
                        new ServerUploadClient(serverResult, imageData).upload();
                    }
                }
                if (serverResult.getZoom() != 0)
                    setZoom = true;
            }
        }
        return null;
    }

    @Override
    protected void onPostExecute(Void aVoid) {
        super.onPostExecute(aVoid);
        if (setZoom) {
            // Turn off server side zoom for now.
            // mPreview.increaseZoom(serverResult.getZoom());
        }
        mServerOverlay.showScreen();
        mServerOverlay.postInvalidate();
    }
}

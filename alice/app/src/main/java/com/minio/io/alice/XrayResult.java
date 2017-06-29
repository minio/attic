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

import org.json.JSONException;
import org.json.JSONObject;

// This class is populated when server sends data. It needs a lot of improvement.
public class XrayResult {
    private String presignedURL;
    private int zoom;
    private Integer frameId;

    JSONObject replyObject = null;
    private static boolean isReply = false;

    public XrayResult(String replyMessage) {
        try {
            replyObject = new JSONObject(replyMessage);
        } catch (JSONException e) {
            e.printStackTrace();
        }

        try {
            setFrameId(replyObject.getInt("FrameId"));
        } catch (JSONException e) {
            e.printStackTrace();
        }

        try {
            setZoom(replyObject.getInt("Zoom"));
        } catch (JSONException e) {
            e.printStackTrace();
        }

        try {
            setPresignedURL(replyObject.getString("URL"));
        } catch (JSONException e) {
            e.printStackTrace();
        }
        isReply = true;
    }

    public static boolean isReply() {
        return isReply;
    }

    public Integer getFrameId() { return frameId; }

    public void setFrameId(int id) { this.frameId = Integer.valueOf(id); }

    public int getZoom() {
        return zoom;
    }

    public void setZoom(int zoom) {
        this.zoom = zoom;
    }

    public void setPresignedURL(String url) {
        this.presignedURL = url;
    }

    public String getPresignedURL() {
        return this.presignedURL;
    }



}

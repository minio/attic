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

import android.hardware.SensorEvent;
import android.util.Log;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.Arrays;

public class SensorRecord {
    private int sensorType;
    private int accuracy;
    private long timestamp;
    private float[] values;
    private String sensorName;

    public SensorRecord(SensorEvent event) {
        this.sensorType = event.sensor.getType();
        this.accuracy = event.accuracy;
        this.timestamp = event.timestamp;
        this.values = new float[event.values.length];
        this.values = event.values;
        this.sensorName = event.sensor.getStringType();
    }

    private JSONObject toJson() {
        JSONObject object = new JSONObject();
        try {
            object.put("sensorName", this.sensorName);
            object.put("sensorType", this.sensorType);
            object.put("timestamp", this.timestamp);
            object.put("accuracy", this.accuracy);
            object.put("values", new JSONArray(Arrays.asList(this.values)));

        } catch (JSONException e) {
            if (XDebug.LOG)
                Log.d(MainActivity.TAG, "JSON exception while assembling sensor record");
        }
        return object;
    }

    public String toString() {
        return this.toJson().toString();
    }

}

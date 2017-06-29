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

import android.location.Location;
import android.util.Log;

import org.json.JSONException;
import org.json.JSONObject;

public class LocationRecord {
    private float accuracy;
    private double altitude;
    private float bearing;
    private double longitude;
    private double latitude;
    private float speed;

    public LocationRecord(Location location) {
        this.accuracy = location.getAccuracy();
        this.altitude = location.getAltitude();
        this.bearing = location.getBearing();
        this.speed = location.getSpeed();
        this.longitude = location.getLongitude();
        this.latitude = location.getLatitude();
    }

    private JSONObject toJson() {
        JSONObject object = new JSONObject();
        try {
            object.put("accuracy", this.accuracy);
            object.put("altitude", this.altitude);
            object.put("bearing", this.bearing);
            object.put("longitude", this.longitude);
            object.put("latitude", this.latitude);
            object.put("speed", this.speed);
        } catch (JSONException e) {
            if (XDebug.LOG)
                Log.d(MainActivity.TAG, "JSON exception while assembling location record");
        }
        return object;
    }

    public String toString() {
        return this.toJson().toString();
    }
}
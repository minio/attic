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
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.hardware.TriggerEvent;
import android.hardware.TriggerEventListener;
import android.util.Log;

import java.util.List;

import static com.minio.io.alice.MainActivity.context;

public class SensorDataLogger implements SensorEventListener {
    private long lastUpdate;
    private static final int MIN_POLLING_DURATION = 1000 * 10; // 1 second

    // for motion detection
    private float mAccel;
    private float mAccelCurrent;
    private float mAccelLast;

    public SensorDataLogger() {
        initAccel();
        // Sensor manager instance
        SensorManager mSensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);

        // Register all available sensors to this listener
        List<Sensor> sensors = mSensorManager.getSensorList(Sensor.TYPE_ALL);
        for (int i = 0; i < sensors.size(); i++) {
            Sensor sensor = sensors.get(i);

            // Request location update if significant motion detected
            if (sensor.getType() == Sensor.TYPE_SIGNIFICANT_MOTION ||
                    (sensor.getType() == Sensor.TYPE_STEP_DETECTOR)) {
                TriggerEventListener mTriggerEventListener = new TriggerEventListener() {
                    @Override
                    public void onTrigger(TriggerEvent event) {
                        if (MainActivity.locationTracker != null)
                            MainActivity.locationTracker.logLastKnownLocation();
                    }
                };
                mSensorManager.requestTriggerSensor(mTriggerEventListener, sensor);
            }

            boolean success = mSensorManager.registerListener(this, sensor, SensorManager.SENSOR_DELAY_NORMAL);
            if (XDebug.LOG)
                Log.d(MainActivity.TAG + "SENSOR", (success ? "REGISTERED: " : "FAILED: ") + sensor.toString());
        }
        lastUpdate = System.currentTimeMillis();
    }

    @Override
    public void onSensorChanged(SensorEvent event) {
        long currentTime = System.currentTimeMillis();
        if ((currentTime - lastUpdate) > MIN_POLLING_DURATION) {
            lastUpdate = System.currentTimeMillis();
            // use accelerometer for motion detection
            if (event.sensor.getType() == Sensor.TYPE_ACCELEROMETER) {
                detectMotionChange(event);
            }
            SensorRecord record = new SensorRecord(event);
            // Disable writing sensor logger for now.
            // write(record.toString());
        }
    }

    private void initAccel() {
        mAccel = 0.0f;
        mAccelCurrent = SensorManager.GRAVITY_EARTH;
        mAccelLast = SensorManager.GRAVITY_EARTH;
    }

    private void detectMotionChange(SensorEvent event) {
        float[] mGravity = event.values.clone();
        //Movement detectioh
        float sensitivityLevel = 0.5f;

        float x = mGravity[0];
        float y = mGravity[1];
        float z = mGravity[2];
        mAccelLast = mAccelCurrent;
        mAccelCurrent = (float) Math.sqrt(x * x + y * y + z * z);
        float delta = mAccelCurrent - mAccelLast;
        mAccel = mAccel * 0.9f + delta;

        //If mobile move any direction then the following condition will become true
        if (mAccel > sensitivityLevel) {
            if (MainActivity.locationTracker != null)
                MainActivity.locationTracker.logLastKnownLocation();

        }

    }

    @Override
    public void onAccuracyChanged(Sensor sensor, int accuracy) {
    }

    public void write(String data) {
        AliceTask vTask = new AliceTask(data);
        vTask.execute();

    }
}

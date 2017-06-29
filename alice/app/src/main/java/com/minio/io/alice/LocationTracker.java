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
import android.location.Location;
import android.location.LocationListener;
import android.location.LocationManager;
import android.os.Bundle;
import android.util.Log;

import java.util.Date;

import static android.content.Context.LOCATION_SERVICE;
import static com.minio.io.alice.MainActivity.context;

public class LocationTracker implements LocationListener {

    private final Context mContext;

    private Location mGPSLocation, mNetworkLocation; // location
    private Location prevLocation;
    private Date mLastUpdateTime; //

    // The minimum distance to change Updates in meters
    private static final long MIN_DISTANCE_CHANGE_FOR_UPDATES = 2; // 2 meters

    // The minimum time between updates in milliseconds
    private static final long MIN_TIME_BW_UPDATES = 1000 * 5; // 5 seconds

    private LocationManager locationManager;
    private boolean isGPSEnabled = false;
    private boolean isNetworkEnabled = false;

    public LocationTracker() {
        this.mContext = context;
        startLocationUpdates();
    }

    protected void startLocationUpdates() {
        try {
            // Declaring a Location Manager
            locationManager = (LocationManager) mContext
                    .getSystemService(LOCATION_SERVICE);

            // getting network status
            isNetworkEnabled = locationManager
                    .isProviderEnabled(LocationManager.NETWORK_PROVIDER);
            // getting GPS status
            isGPSEnabled = locationManager
                    .isProviderEnabled(LocationManager.GPS_PROVIDER);

            boolean canGetLocation = false;

            // if GPS Enabled get lat/long using GPS Services
            if (isGPSEnabled) {
                canGetLocation = true;
                if (mGPSLocation == null) {
                    locationManager.requestLocationUpdates(
                            LocationManager.GPS_PROVIDER,
                            MIN_TIME_BW_UPDATES,
                            MIN_DISTANCE_CHANGE_FOR_UPDATES, this);
                    if (XDebug.LOG)
                        Log.d(MainActivity.TAG, "GPS Enabled");

                }
            }
            // First get location from Network Provider
            if (isNetworkEnabled) {
                canGetLocation = true;
                locationManager.requestLocationUpdates(
                        LocationManager.NETWORK_PROVIDER,
                        MIN_TIME_BW_UPDATES,
                        MIN_DISTANCE_CHANGE_FOR_UPDATES, this);
                if (XDebug.LOG)
                    Log.d(MainActivity.TAG, "Network");

            }
            if (!canGetLocation) {
                // no network provider is enabled
                if (XDebug.LOG)
                    Log.d(MainActivity.TAG, "GPS and Network disabled");
            }
        } catch (SecurityException e) {
            Log.e(MainActivity.TAG, "Location service error");
        }
    }

    public void logLastKnownLocation() {
        try {
            if (locationManager == null) {
                startLocationUpdates();
            } else if (isGPSEnabled) {
                mGPSLocation = locationManager.getLastKnownLocation(LocationManager.GPS_PROVIDER);
                write(mGPSLocation);
            } else if (isNetworkEnabled) {
                mNetworkLocation = locationManager.getLastKnownLocation(LocationManager.NETWORK_PROVIDER);
                write(mNetworkLocation);
            }
        } catch (SecurityException e) {
            Log.e(MainActivity.TAG, "Location update error");
        }

    }

    @Override
    public void onLocationChanged(Location location) {
        if (XDebug.LOG)
            Log.d(MainActivity.TAG, location.toString());
        // Disable sending local change for now.
        // write(location);
    }

    protected void stopLocationUpdates() {
        if (this.locationManager != null) {
            try {
                this.locationManager.removeUpdates(LocationTracker.this);
            } catch (SecurityException e) {
                Log.d(MainActivity.TAG, "GPS error");
            }
        }
    }

    /**
     * Determines whether one Location reading is better than the current Location fix
     *
     * @param location            The new Location that you want to evaluate
     * @param currentBestLocation The current Location fix, to which you want to compare the new one
     */
    protected boolean isBetterLocation(Location location, Location currentBestLocation) {
        if (currentBestLocation == null) {
            // A new location is always better than no location
            return true;
        }

        // Check whether the new location fix is newer or older
        long timeDelta = location.getTime() - currentBestLocation.getTime();
        boolean isSignificantlyNewer = timeDelta > MIN_TIME_BW_UPDATES;
        boolean isSignificantlyOlder = timeDelta < -MIN_TIME_BW_UPDATES;
        boolean isNewer = timeDelta > 0;

        // If it's been more than two minutes since the current location, use the new location
        // because the user has likely moved
        if (isSignificantlyNewer) {
            return true;
            // If the new location is more than two minutes older, it must be worse
        } else if (isSignificantlyOlder) {
            return false;
        }

        // Check whether the new location fix is more or less accurate
        int accuracyDelta = (int) (location.getAccuracy() - currentBestLocation.getAccuracy());
        boolean isLessAccurate = accuracyDelta > 0;
        boolean isMoreAccurate = accuracyDelta < 0;
        boolean isSignificantlyLessAccurate = accuracyDelta > 200;

        // Check if the old and new location are from the same provider
        boolean isFromSameProvider = isSameProvider(location.getProvider(),
                currentBestLocation.getProvider());

        // Determine location quality using a combination of timeliness and accuracy
        if (isMoreAccurate) {
            return true;
        } else if (isNewer && !isLessAccurate) {
            return true;
        } else if (isNewer && !isSignificantlyLessAccurate && isFromSameProvider) {
            return true;
        }
        return false;
    }

    /**
     * Checks whether two providers are the same
     */
    private boolean isSameProvider(String provider1, String provider2) {
        if (provider1 == null) {
            return provider2 == null;
        }
        return provider1.equals(provider2);
    }

    @Override
    public void onProviderDisabled(String provider) {
    }

    @Override
    public void onProviderEnabled(String provider) {
    }

    @Override
    public void onStatusChanged(String provider, int status, Bundle extras) {
    }

    // Determine if location change meets threshold
    private boolean isLocationChangeThreshold(Location location) {
        //if first location fetch
        if (prevLocation == null) {
            return true;
        }
        return (Math.abs(location.distanceTo(prevLocation)) > MIN_DISTANCE_CHANGE_FOR_UPDATES);


    }

    // write location record if more accurate or significant movement
    public void write(Location location) {
        if (isBetterLocation(location, prevLocation) || isLocationChangeThreshold(location)) {
            LocationRecord record = new LocationRecord(location);
            prevLocation = location;
            AliceTask vTask = new AliceTask(record.toString());
            vTask.execute();

        }
    }

}

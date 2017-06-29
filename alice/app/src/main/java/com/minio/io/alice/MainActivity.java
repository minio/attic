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

import android.Manifest;
import android.app.Activity;
import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.net.Uri;
import android.os.Bundle;
import android.os.SystemClock;
import android.provider.Settings;
import android.support.annotation.NonNull;
import android.support.design.widget.Snackbar;
import android.support.v4.app.ActivityCompat;
import android.support.v4.content.ContextCompat;
import android.util.Log;
import android.view.GestureDetector;
import android.view.MotionEvent;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;
import android.widget.SeekBar;
import android.widget.Toast;

import com.google.android.gms.vision.MultiProcessor;
import com.google.android.gms.vision.Tracker;
import com.google.android.gms.vision.barcode.Barcode;
import com.google.android.gms.vision.face.Face;

import net.hockeyapp.android.CrashManager;
import net.hockeyapp.android.UpdateManager;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.UUID;

public class MainActivity extends Activity  implements PreviewCallback {

    public static ClientWebSocket webSocket = null;
    public static Context context;
    public static String TAG = "__ALICE__";

    private static final int REQUEST_VIDEO_PERMISSIONS = 1;
    private boolean hasVideoPermission = false;
    private boolean hasLocationPermission = false;

    boolean mServiceBound = false;
    boolean serverThreadStarted = false;
    boolean serverThreadRunning = false;
    private static LinkedList<XrayResult> mServerReplyQueue;
    private static final int MAX_BUFFER = 50;

    public static LocationTracker locationTracker;

    private static final String[] VIDEO_PERMISSIONS = {
            Manifest.permission.CAMERA,
            Manifest.permission.RECORD_AUDIO,
            Manifest.permission.ACCESS_FINE_LOCATION,
    };

    GestureDetector gestureDetector;

    private CameraSourcePreview mPreview;
    private GraphicOverlay mGraphicOverlay;

    private static int mCameraId = CameraSource.CAMERA_FACING_BACK;

    private CameraDeviceManager cameraManager;

    Thread frameHandlerThread;
    FrameHandler frameHandler;
    public static boolean frameHandlerStarted = false;
    ServerHandler serverhandler;
    private ServerResponseHandler serverResponseHandler;
    private Thread serverResponseThread;

    public static boolean isAliceAwake = false;
    public static long prevFaceDetectionAt = 0;
    private long currentTime;
    private static final int ELAPSED_DURATION = 60000;  // 1 minute

    public static final String CLIENT_UUID = UUID.randomUUID().toString();

    public MainActivity() {
        gestureDetector = new GestureDetector(context, new GestureListener());
        if (XDebug.LOG)
            Log.i(MainActivity.TAG, "Instantiated new " + this.getClass());
    }

    @Override
    public boolean onTouchEvent(MotionEvent e) {
        return gestureDetector.onTouchEvent(e);
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        context = this;
        if (webSocket == null) {
            webSocket = new ClientWebSocket();
            Log.i(MainActivity.TAG, "About to connect to WS");
            webSocket.connect(context);
        }

        // Init media writers and location,sensor trackers
        serverhandler = new ServerHandler(context);

        // Spawn thread for handler for server response to Alice
        serverResponseHandler = new ServerResponseHandler();
        serverResponseThread = new Thread(serverResponseHandler);
        mServerReplyQueue = new LinkedList<XrayResult>();

        //Spawn thread for frame handler
        initFrameHandler();

        requestWindowFeature(Window.FEATURE_NO_TITLE);
        this.getWindow().setFlags(WindowManager.LayoutParams.FLAG_FULLSCREEN, WindowManager.LayoutParams.FLAG_FULLSCREEN);

        setContentView(R.layout.activity_main);

        mPreview = (CameraSourcePreview) findViewById(R.id.ZoomCameraView);
        mPreview.setZoomControl((SeekBar) findViewById(R.id.CameraZoomControls));
        mGraphicOverlay = (GraphicOverlay) findViewById(R.id.faceOverlay);

        if (cameraManager == null) {
            cameraManager = new CameraDeviceManager(context,
                    new OcrDetectorProcessor(mGraphicOverlay),
                    new GraphicFaceTrackerFactory(),
                    new BarcodeTrackerFactory(),
                    frameHandler);
        }

        // Check for the camera permission before accessing the camera.  If the
        // permission is not granted yet, request permission.
        if (!checkSelfPermission(VIDEO_PERMISSIONS))
            cameraManager.createCameraSource(mCameraId);
        else
            requestVideoPermission();

        // Check for updates on hockey.
        checkForUpdates();
    }

    protected ServerHandler getServerHandler() {
        return this.serverhandler;
    }

    protected FrameHandler getFrameHandler() {
        return this.frameHandler;
    }
    protected Thread getFrameHandlerThread() {
        return this.frameHandlerThread;
    }

    protected CameraSourcePreview getPreview() { return this.mPreview; }

    @Override
    public void onPause() {
        super.onPause();

        serverThreadRunning = false;

        cameraManager.pauseCameraResource();
        mPreview.stop();
        serverhandler.stop();
        unregisterManagers();
    }

    //Spawns a new Framehandler thread
    private void initFrameHandler() {
        if (frameHandlerThread == null) {
            frameHandler = new FrameHandler();
            frameHandlerThread = new Thread(frameHandler);
        }
    }

    @Override
    public void onResume() {

        super.onResume();

        if (webSocket == null) {
            webSocket = new ClientWebSocket();
            webSocket.connect(context);
        }
        serverThreadRunning = true;
        cameraManager.startCameraSource();
        serverhandler.start();
        initFrameHandler();

        if (!serverThreadStarted) {
            serverThreadStarted = true;
            serverResponseThread.start();
        }

        checkForCrashes();
    }

    public void onDestroy() {

        super.onDestroy();

        frameHandlerThread = null;
        serverhandler.stop();
        serverResponseThread = null;
        serverThreadRunning = false;

        cameraManager.releaseCameraResource();
        unregisterManagers();
    }

    // Use Hockey Framework to collect crash reports on clients.
    private void checkForCrashes() {
        CrashManager.register(this);
    }

    // Hockey App Distribution
    private void checkForUpdates() {
        // Remove this for store builds!
        // UpdateManager.register(this);

    }

    private void unregisterManagers() {
        UpdateManager.unregister();
    }

    // Private class to handle touch events on camera.
    private class GestureListener extends GestureDetector.SimpleOnGestureListener {

        @Override
        public boolean onDown(MotionEvent e) {
            return false;
        }

        // event when double tap occurs
        @Override
        public boolean onDoubleTap(MotionEvent e) {
            float x = e.getX();
            float y = e.getY();
            if (XDebug.LOG)
                Log.d(MainActivity.TAG, "Tapped at: (" + x + "," + y + ")");
            swapCamera();
            return true;
        }

        @Override
        public void onLongPress(MotionEvent event) {
            // triggers after onDown only for long press
            if(XDebug.LOG)
                Log.i(MainActivity.TAG, "Long Press");

            super.onLongPress(event);

        }
    }


    // Upon double tap, swap front  and back cameras
    public void swapCamera() {

        mCameraId = (mCameraId ==  CameraSource.CAMERA_FACING_FRONT) ? CameraSource.CAMERA_FACING_BACK : CameraSource.CAMERA_FACING_FRONT;
        cameraManager.swapCameraSource();
    }

    /**
     * Factory for creating a face tracker to be associated with a new face.  The multiprocessor
     * uses this factory to create face trackers as needed -- one for each individual.
     */
    private class GraphicFaceTrackerFactory implements MultiProcessor.Factory<Face> {
        @Override
        public Tracker<Face> create(Face face) {
            return new GraphicFaceTracker(mGraphicOverlay);
        }
    }

    /**
     * Factory for creating a tracker and associated graphic to be associated with a new barcode.  The
     * multi-processor uses this factory to create barcode trackers as needed -- one for each barcode.
     */
    class BarcodeTrackerFactory implements MultiProcessor.Factory<Barcode> {
        @Override
        public Tracker<Barcode> create(Barcode barcode) {
            BarcodeGraphic graphic = new BarcodeGraphic(mGraphicOverlay);
            return new BarcodeGraphicTracker(mGraphicOverlay, graphic);
        }

    }

    // Callback for CameraDeviceManager to start preview
    public void startPreview(CameraSource cameraSource) throws IOException {
        mPreview.start(cameraSource, mGraphicOverlay);
    }

    private class ServerResponseHandler implements Runnable {

        ServerResponseTask stask = null;

        public ServerResponseHandler() {
        }

        @Override
        public void run() {
            while (serverThreadRunning ) {
                XrayResult serverReply = dequeueServerReply();
                manageAliceDisplay();
                if (serverReply != null) {
                    stask = new ServerResponseTask(context, serverReply, mGraphicOverlay, mPreview);
                    stask.execute();
                }
            }
        }
    }

    // Enqueue server messages.
    public static void enqueueServerReply(XrayResult serverReply) {
        if (mServerReplyQueue == null)
            return;
        synchronized (mServerReplyQueue) {
            if (mServerReplyQueue.size() == MAX_BUFFER) {
                mServerReplyQueue.poll();
            }
            mServerReplyQueue.add(serverReply);
        }
    }
    //Poll mServerReplyQueue and process messages
    public XrayResult dequeueServerReply() {
        synchronized (mServerReplyQueue) {
            return mServerReplyQueue.poll();
        }
    }

    private void manageAliceDisplay() {
        currentTime = SystemClock.elapsedRealtime();
        long timeElapsed = currentTime - prevFaceDetectionAt;
        if (isAliceAwake && (timeElapsed > ELAPSED_DURATION)) {
            isAliceAwake = false;
            resetZoom(mPreview);
        }

        //Wake up Alice if currently invisible
        if (isAliceAwake && !mPreview.isShown())
            toggleDisplay(mPreview,View.VISIBLE);

        //Blank display if Alice has not detected any face in ELAPSED_DURATION
        if (!isAliceAwake && mPreview.isShown())
            toggleDisplay(mPreview,View.INVISIBLE);
    }

    private void toggleDisplay(final View view, final int visibility) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                view.setVisibility(visibility);
            }
        });
    }

    private void resetZoom(final CameraSourcePreview preview) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                preview.resetZoom();
            }
        });
    }
    // Get Video Permissions

    /**
     * Gets whether you should show UI with rationale for requesting permissions.
     *
     * @param permissions The permissions your app wants to request.
     * @return Whether you can show permission rationale UI.
     */
    private boolean shouldShowRequestPermissionRationale(String[] permissions) {
        boolean show =  false;
        for (String permission : permissions) {
            if (ActivityCompat.shouldShowRequestPermissionRationale(this, permission)) {
                show =  true;
            }
        }
        return show;
    }

    private boolean checkSelfPermission(String[] permissions) {
        for (String permission : permissions) {
            if (ContextCompat.checkSelfPermission(this, permission)
                    != PackageManager.PERMISSION_GRANTED)
                return true;
        }
        return false;
    }

    private void requestVideoPermission() {

        if (shouldShowRequestPermissionRationale(VIDEO_PERMISSIONS)) {
            // Show an explanation to the user *asynchronously* -- don't block
            // this thread waiting for the user's response! After the user
            // sees the explanation, try again to request the permission.
            ActivityCompat.requestPermissions(this,
                    VIDEO_PERMISSIONS,
                    REQUEST_VIDEO_PERMISSIONS);

            //showDialog(VIDEO_PERMISSIONS);
        } else {
            // No explanation needed, we can request the permission.
            ActivityCompat.requestPermissions(this,
                    VIDEO_PERMISSIONS,
                    REQUEST_VIDEO_PERMISSIONS);

            // REQUEST_VIDEO_PERMISSIONS is an
            // app-defined int constant. The callback method gets the
            // result of the request.
        }



    }
    // Find permissions that were not granted and return as an ArrayList
    private ArrayList<String> getPendingPermissions(int[] grantResults, String permissions[]) {
        HashMap<String,Integer> perms = new HashMap();
        ArrayList<String> pendingPermissions = new ArrayList<String>();
        for (int i = 0; i < permissions.length; i++) {
            perms.put(permissions[i],grantResults[i]);
            if (permissions[i] == Manifest.permission.ACCESS_FINE_LOCATION && grantResults[i] == PackageManager.PERMISSION_GRANTED)
                this.hasLocationPermission = true;
            if (permissions[i] == Manifest.permission.CAMERA  && grantResults[i] == PackageManager.PERMISSION_GRANTED)
                this.hasVideoPermission = true;
        }

        for (int i = 0; i < VIDEO_PERMISSIONS.length; i++) {
            if (grantResults.length == VIDEO_PERMISSIONS.length) {
                if (perms.get(VIDEO_PERMISSIONS[i]) != PackageManager.PERMISSION_GRANTED) {
                    pendingPermissions.add(VIDEO_PERMISSIONS[i]);
                }
            } else {
                pendingPermissions.add(VIDEO_PERMISSIONS[i]);
            }
        }
        return pendingPermissions;
    }

    // Callback with the request from calling requestPermissions(...)
    @Override
    public void onRequestPermissionsResult(int requestCode,
                                           @NonNull String permissions[],
                                           @NonNull int[] grantResults) {
        // Make sure it's our original REQUEST_VIDEO_PERMISSIONS request.
        if (requestCode == REQUEST_VIDEO_PERMISSIONS) {
            ArrayList<String> permissionsNeededYet = getPendingPermissions(grantResults,permissions);
            if (permissionsNeededYet.size() == 0){
                // All permissions granted - allow camera access.
                cameraManager.createCameraSource(mCameraId);
                return;

            } else {
                // showRationale = false if user clicks Never Ask Again, otherwise true.
                boolean showRationale = shouldShowRequestPermissionRationale(permissionsNeededYet.toArray(new String[0]));

                if (!showRationale) {
                    Toast.makeText(this, "Video permission denied.Enable camera and location preferences on the App settings", Toast.LENGTH_SHORT).show();
                    finish();
                } else {
                    Toast.makeText(this, "Alice needs camera, microphone and location enabled to start tracking. Alice will now exit.", Toast.LENGTH_SHORT).show();
                    startActivity(new Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS, Uri.parse("package:" + BuildConfig.APPLICATION_ID)));

                    finish();
                }
            }
        } else {
            super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        }
    }

    // Show dialog to request for permissions.
    void showDialog(final String permissions[]) {
        final Activity thisActivity = this;
        View.OnClickListener listener = new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                ActivityCompat.requestPermissions(thisActivity, permissions, REQUEST_VIDEO_PERMISSIONS);
            }
        };

        Snackbar.make(mGraphicOverlay, R.string.permission_camera_rationale,
                Snackbar.LENGTH_INDEFINITE)
                .setAction(R.string.ok, listener)
                .show();
    }
}

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

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.util.Log;

/**
 * Incoming Reciever reacts to messages coming back from the XRay Server. It broadcasts Xray server's response.
 */

public class IncomingReceiver extends BroadcastReceiver {

    public IncomingReceiver() {}

    @Override
    public void onReceive(Context context, Intent intent) {
        String msg = intent.getStringExtra(String.valueOf(R.string.xray_broadcast));
        if (msg != null) {
            if(XDebug.LOG)
                Log.i(MainActivity.TAG, msg);
            MainActivity.enqueueServerReply(new XrayResult(msg));
        }
    }
}

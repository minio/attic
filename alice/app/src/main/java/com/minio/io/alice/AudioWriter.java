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
import android.media.AudioFormat;
import android.media.AudioRecord;
import android.media.MediaRecorder;

public class AudioWriter {

    private Context context;

    private final int CHANNEL_CONFIG = AudioFormat.CHANNEL_IN_MONO;
    private final int AUDIO_ENCODING = AudioFormat.ENCODING_PCM_16BIT;
    private final int SAMPLE_RATE = 44100;

    private  int BUFFER_SIZE  ;
    private volatile Thread recordingThread;
    private AudioRecord audioRecord;

    // remove this flag when server can handle audio
    private boolean sendAudio;

    private boolean recording;

    public AudioWriter(Context context) {
        BUFFER_SIZE = AudioRecord.getMinBufferSize(SAMPLE_RATE, CHANNEL_CONFIG, AUDIO_ENCODING);
        this.context = context;
        startRecording();
    }


    public AudioWriter(Context context,boolean audioFlag) {
        this(context);
        this.sendAudio = audioFlag;
    }


    public boolean isRecording() {
        return (audioRecord != null) && (audioRecord.getRecordingState() == AudioRecord.RECORDSTATE_RECORDING);
    }

    private void prepareAudio() {
        if (audioRecord == null) {
            try {
                audioRecord = new AudioRecord(MediaRecorder.AudioSource.MIC,
                    SAMPLE_RATE,
                    CHANNEL_CONFIG,
                    AUDIO_ENCODING, BUFFER_SIZE);
             } catch (IllegalArgumentException e) {
                audioRecord = null;
             }
        }
    }

    public void startRecording() {

        try {
            if (audioRecord == null){
                prepareAudio();
            }
            audioRecord.startRecording();
            recording = true;
        } catch (IllegalStateException e) {
            e.printStackTrace();
        }

        if (sendAudio == false)
            return;

        recordingThread = new Thread(new Runnable() {
            @Override
            public void run() {
                //Write the output audio in byte
                byte baudioBuffer[] = new byte[BUFFER_SIZE * 2];

                while (recording) {
                    // gets the voice output from microphone to byte format
                    audioRecord.read(baudioBuffer, 0, BUFFER_SIZE * 2);
                    AliceTask vTask = new AliceTask(baudioBuffer);
                    vTask.execute();
                }

            }
        });

        recordingThread.start();
    }

    public void stopRecording() {
        recordingThread =  null;
        recording = false;
    }

}



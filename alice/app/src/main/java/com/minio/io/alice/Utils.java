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

import android.graphics.Bitmap;
import android.graphics.Matrix;

class Utils {

    // Rotate bitmap depending on rotation parameter.
    // For rotation of
    // - 1 its 90*
    // - 2 its 180*
    // - 3 its 270*
    // resulting bitmap is of the right requested orientation.
    public static Bitmap rotateBitmap(Bitmap bitmap, int rotation) {
        // Allocate a new matrix to be used to rotate the bitmap
        // to correct orientation.
        Matrix matrix = new Matrix();
        switch (rotation) {
            case 1:
                matrix.setRotate(90);
                break;
            case 2:
                matrix.setRotate(180);
                break;
            case 3:
                matrix.setRotate(270);
                break;
            default:
                break;
        }

        // Create a new bitmap with correct orientation.
        return Bitmap.createBitmap(bitmap, 0, 0, bitmap.getWidth(), bitmap.getHeight(), matrix, true);
    }
}
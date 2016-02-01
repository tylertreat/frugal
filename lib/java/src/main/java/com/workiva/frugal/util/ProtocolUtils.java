package com.workiva.frugal.util;

import java.nio.charset.Charset;

public class ProtocolUtils {

    public static int readInt(byte[] buff, int offset) {
        return ((buff[offset] & 0xff) << 24) |
                ((buff[offset + 1] & 0xff) << 16) |
                ((buff[offset + 2] & 0xff) << 8) |
                (buff[offset + 3] & 0xff);
    }

    public static void writeInt(int i, byte[] buff, int offset) {
        buff[offset] = (byte) (0xff & (i >> 24));
        buff[offset + 1] = (byte) (0xff & (i >> 16));
        buff[offset + 2] = (byte) (0xff & (i >> 8));
        buff[offset + 3] = (byte) (0xff & (i));
    }

    public static void writeString(String s, byte[] buff, int offset) {
        byte[] strBytes = Charset.forName("UTF-8").encode(s).array();
        System.arraycopy(strBytes, 0, buff, offset, s.length());
    }

}

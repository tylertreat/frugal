package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import java.util.Arrays;

import static org.junit.Assert.*;

public class FBoundedMemoryBufferTest {

    private FBoundedMemoryBuffer buffer;

    @Before
    public void setUp() {
        buffer = new FBoundedMemoryBuffer(10);
    }

    @Test
    public void testWrite() throws TTransportException {
        buffer.write("foo".getBytes());
        assertArrayEquals("foo".getBytes(), Arrays.copyOfRange(buffer.getArray(), 0, buffer.length()));
    }

    @Test
    public void testWriteLen() throws TTransportException {
        buffer.write("foooooooo".getBytes(), 0, 3);
        assertArrayEquals("foo".getBytes(), Arrays.copyOfRange(buffer.getArray(), 0, buffer.length()));
    }

    @Test(expected = FMessageSizeException.class)
    public void testWrite_sizeException() throws TTransportException {
        assertEquals(0, buffer.length());
        buffer.write(new byte[11]);
        assertEquals(0, buffer.length());
    }

    @Test(expected = FMessageSizeException.class)
    public void testWriteLen_size_Exception() throws TTransportException {
        assertEquals(0, buffer.length());
        buffer.write(new byte[11], 0, 10);
        assertEquals(10, buffer.length());
        buffer.write(new byte[11], 10, 1);
        assertEquals(0, buffer.length());
    }

}

package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import org.apache.thrift.transport.TMemoryBuffer;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

/**
 * An implementation of TTransport using a bounded memory buffer. Writes which cause the buffer to exceed its size
 * throw an FMessageSizeException.
 */
public class FBoundedMemoryBuffer extends TTransport {

    private TMemoryBuffer buffer;
    private final int limit;

    public FBoundedMemoryBuffer(int size) {
        buffer = new TMemoryBuffer(size);
        limit = size;
    }

    @Override
    public boolean isOpen() {
        return buffer.isOpen();
    }

    @Override
    public void open() throws TTransportException {
        buffer.open();
    }

    @Override
    public void close() {
        buffer.close();
    }

    @Override
    public int read(byte[] buf, int off, int len) throws TTransportException {
        return buffer.read(buf, off, len);
    }

    @Override
    public void write(byte[] buf) throws TTransportException {
        if (buffer.length() + buf.length > limit) {
            buffer = new TMemoryBuffer(limit);
            throw new FMessageSizeException(String.format("Buffer size reached (%d)", limit));
        }
        buffer.write(buf);
    }

    @Override
    public void write(byte[] buf, int off, int len) throws TTransportException {
        if (buffer.length() + len > limit) {
            buffer = new TMemoryBuffer(limit);
            throw new FMessageSizeException(String.format("Buffer size reached (%d)", limit));
        }
        buffer.write(buf, off, len);
    }

    public int length() {
        return buffer.length();
    }

    public byte[] getArray() {
        return buffer.getArray();
    }

}

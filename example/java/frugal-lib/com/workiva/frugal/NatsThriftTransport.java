package com.workiva.frugal;

import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.nats.Connection;
import org.nats.MsgHandler;

import java.io.IOException;
import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.nio.ByteBuffer;
import java.util.Arrays;

/**
 * NatsThriftTransport is an extension of thrift.TTransport exclusively used for
 * pub/sub via NATS.
 */
public class NatsThriftTransport extends TTransport {

    // NATS limits messages to 1MB.
    private static final int NATS_MAX_MESSAGE_SIZE = 1024 * 1024;
    private static final int UNSUBSCRIBED = -1;

    private Connection conn;
    private PipedOutputStream writer;
    private PipedInputStream reader;
    private ByteBuffer writeBuffer;
    private int sub;
    private String subject;

    public NatsThriftTransport(Connection conn) {
        this.conn = conn;
        writeBuffer = ByteBuffer.allocate(NATS_MAX_MESSAGE_SIZE);
        sub = UNSUBSCRIBED;
    }

    @Override
    public boolean isOpen() {
        return sub != UNSUBSCRIBED;
    }

    @Override
    public void open() throws TTransportException {
        try {
            writer = new PipedOutputStream();
            reader = new PipedInputStream(writer);
        } catch (IOException e) {
            throw new TTransportException(TTransportException.UNKNOWN, e);
        }

        try {
            sub = conn.subscribe(subject, new MsgHandler() {
                @Override
                public void execute(byte[] msg, String reply, String subject) {
                    try {
                        writer.write(msg);
                    } catch (IOException e) {
                        // TODO: What do we do here?
                        e.printStackTrace();
                    }
                }
            });
        } catch (IOException e) {
            throw new TTransportException(TTransportException.UNKNOWN, e);
        }
    }

    @Override
    public void close() {
        if (!isOpen()) {
            return;
        }
        try {
            conn.unsubscribe(sub);
        } catch (IOException e) {
            // TODO: What do we do here?
            e.printStackTrace();
        }
        sub = UNSUBSCRIBED;
        try {
            writer.close();
        } catch (IOException e) {
            // TODO: What do we do here?
            e.printStackTrace();
        }
    }

    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        try {
            return reader.read(bytes, off, len);
        } catch (IOException e) {
            throw new TTransportException(TTransportException.UNKNOWN, e);
        }
    }

    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (bytes.length + writeBuffer.position() > writeBuffer.capacity()) {
            writeBuffer.clear();
            throw new TTransportException(TTransportException.UNKNOWN,
                    "Message is too large");
        }
        writeBuffer.put(bytes, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        byte[] data = new byte[writeBuffer.position()];
        writeBuffer.flip();
        writeBuffer.get(data);
        if (data.length == 0) {
            return;
        }
        try {
            conn.publish(subject, null, data, null);
        } catch (IOException e) {
            throw new TTransportException(TTransportException.UNKNOWN, e);
        } finally {
            writeBuffer.clear();
        }
    }

    public void setSubject(String subject) {
        this.subject = subject;
    }

}

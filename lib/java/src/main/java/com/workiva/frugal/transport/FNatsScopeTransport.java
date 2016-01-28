package com.workiva.frugal.transport;

import com.workiva.frugal.FException;
import io.nats.client.*;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.nio.ByteBuffer;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.locks.ReentrantLock;
import java.util.logging.Logger;

/**
 * FNatsScopeTransport implements FScopeTransport by using NATS as the pub/sub message broker. Messages are limited to
 * 1MB in size.
 */
public class FNatsScopeTransport extends FScopeTransport {

    private final Connection conn;
    private String subject;
    private PipedOutputStream writer;
    private PipedInputStream reader;
    private ByteBuffer writeBuffer;
    private Subscription sub;
    private boolean pull;
    private AtomicBoolean isOpen = new AtomicBoolean(false);
    private final ReentrantLock lock;

    private static Logger LOGGER = Logger.getLogger(FNatsScopeTransport.class.getName());

    protected FNatsScopeTransport(Connection conn) {
        this.conn = conn;
        this.lock = new ReentrantLock();
    }

    public static class Factory implements FScopeTransportFactory {

        private Connection conn;

        public Factory(Connection conn) {
            this.conn = conn;
        }

        /**
         * Get a new FScopeTransport instance.
         *
         * @return A new FScopeTransport instance.
         */
        public FNatsScopeTransport getTransport() {
            return new FNatsScopeTransport(this.conn);
        }
    }

    @Override
    public void lockTopic(String topic) throws TException {
        if (pull) {
            throw new FException("subscriber cannot lock topic");
        }
        lock.lock();
        subject = topic;
    }

    @Override
    public void unlockTopic() throws TException {
        if (pull) {
            throw new FException("subscriber cannot unlock topic");
        }
        lock.unlock();
        subject = "";
    }

    @Override
    public void subscribe(String topic) throws TException {
        pull = true;
        subject = topic;
        open();
    }

    @Override
    public boolean isOpen() {
        return conn.getState() == Constants.ConnState.CONNECTED && isOpen.get();
    }

    @Override
    public void open() throws TTransportException {
        if (isOpen()) {
            return;
        }

        isOpen.set(true);

        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw new TTransportException(TTransportException.NOT_OPEN,
                    "NATS not connected, has status " + conn.getState());
        }

        if (!pull) {
            writeBuffer = ByteBuffer.allocate(TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE);
            return;
        }

        if ("".equals(subject)) {
            throw new TTransportException("Subject cannot be empty.");
        }

        try {
            writer = new PipedOutputStream();
            reader = new PipedInputStream(writer);
        } catch (IOException e) {
            throw new TTransportException(e);
        }

        sub = conn.subscribe(subject, new MessageHandler() {
            @Override
            public void onMessage(Message msg) {
                try {
                    writer.write(msg.getData());
                    writer.flush();
                } catch (IOException e) {
                    // pipe is closed, nothing to do.
                }
            }
        });
    }

    @Override
    public void close() {
        if (!isOpen()) {
            return;
        }

        isOpen.set(false);

        if (!pull) {
            return;
        }
        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warning("could not unsubscribe from subscription. " + e.getMessage());
        }
        sub = null;
        try {
            writer.close();
        } catch (IOException e) {
            LOGGER.warning("could not close write buffer. " + e.getMessage());
        }
        writer = null;
        reader = null;
    }

    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }
        try {
            int bytesRead = reader.read(bytes, off, len);
            if (bytesRead < 0) {
                throw new TTransportException(TTransportException.END_OF_FILE);
            }
            return bytesRead;
        } catch (IOException e) {
            throw new TTransportException(TTransportException.END_OF_FILE, e);
        }
    }

    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN, "NATS transport not open");
        }
        if (writeBuffer.remaining() < len) {
            writeBuffer.clear();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes",
                            TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE,
                            len + TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE - writeBuffer.remaining()));
        }
        writeBuffer.put(bytes, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.NOT_OPEN, "NATS transport not open");
        }
        byte[] data = new byte[writeBuffer.position()];
        writeBuffer.flip();
        writeBuffer.get(data);
        if (data.length == 0) {
            return;
        }
        if (data.length > TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE) {
            throw new FMessageSizeException(String.format(
                    "Message exceeds %d bytes, was %d bytes",
                    TNatsServiceTransport.NATS_MAX_MESSAGE_SIZE, data.length));
        }
        conn.publish(subject, data);
        writeBuffer.clear();
    }
}

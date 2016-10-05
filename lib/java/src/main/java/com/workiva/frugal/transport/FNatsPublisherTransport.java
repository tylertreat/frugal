package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.util.ProtocolUtils;
import io.nats.client.Connection;
import io.nats.client.Constants;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.concurrent.locks.ReentrantLock;

import static com.workiva.frugal.transport.FNatsTransport.FRUGAL_PREFIX;
import static com.workiva.frugal.transport.FNatsTransport.NATS_MAX_MESSAGE_SIZE;
import static com.workiva.frugal.transport.FNatsTransport.getClosedConditionException;

/**
 * FNatsPublisherTransport implements FPublisherTransport by using NATS as the pub/sub message broker.
 * Messages are limited to 1MB in size.
 */
public class FNatsPublisherTransport extends FPublisherTransport {
    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsScopeTransport.class);

    private final Connection conn;
    protected String subject;
    private ByteBuffer writeBuffer;
    protected boolean isOpen;
    private final ReentrantLock lock;


    /**
     * Creates a new FNatsPublisherTransport which is used for publishing.
     *
     * @param conn  NATS connection
     */
    protected FNatsPublisherTransport(Connection conn) {
        this.conn = conn;
        this.lock = new ReentrantLock();
    }

    /**
     * An FPublisherTransportFactory implementation which creates FPublisherTransports backed by NATS.
     */
    public static class Factory implements FPublisherTransportFactory {

        private final Connection conn;

        /**
         * Creates a NATS FPublisherTransportFactory using the provided NATS connection.
         *
         * @param conn NATS connection
         */
        public Factory(Connection conn) {
            this.conn = conn;
        }

        /**
         * Get a new FPublisherTransport instance.
         *
         * @return A new FPublisherTransport instance.
         */
        public FPublisherTransport getTransport() {
            return new FNatsPublisherTransport(this.conn);
        }
    }

    @Override
    public void lockTopic(String topic) throws TException {
        lock.lock();
        subject = topic;
    }

    @Override
    public void unlockTopic() throws TException {
        lock.unlock();
        subject = "";
    }

    @Override
    public synchronized boolean isOpen() {
        return conn.getState() == Constants.ConnState.CONNECTED && isOpen;
    }

    @Override
    public synchronized void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw new TTransportException(TTransportException.NOT_OPEN,
                    "NATS not connected, has status " + conn.getState());
        }
        if (isOpen) {
            throw new TTransportException(TTransportException.ALREADY_OPEN, "NATS transport already open");
        }

        writeBuffer = ByteBuffer.allocate(NATS_MAX_MESSAGE_SIZE);
        isOpen = true;
    }

    @Override
    public synchronized void close() {
        isOpen = false;
    }

    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        // Include 4 bytes for frame size.
        if (writeBuffer.remaining() < len + 4) {
            int size = 4 + len + NATS_MAX_MESSAGE_SIZE - writeBuffer.remaining();
            writeBuffer.clear();
            throw new FMessageSizeException(
                    String.format("Message exceeds %d bytes, was %d bytes",
                            NATS_MAX_MESSAGE_SIZE, size));
        }
        writeBuffer.put(bytes, off, len);
    }

    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn.getState(), "flush:");
        }
        byte[] data = new byte[writeBuffer.position()];
        writeBuffer.flip();
        writeBuffer.get(data);
        if (data.length == 0) {
            return;
        }
        byte[] frame = new byte[data.length + 4];
        ProtocolUtils.writeInt(data.length, frame, 0);
        System.arraycopy(data, 0, frame, 4, data.length);
        try {
            conn.publish(getFormattedSubject(), frame);
        } catch (IOException e) {
            throw new TTransportException("flush: unable to publish data: " + e.getMessage());
        }
        writeBuffer.clear();
    }

    private String getFormattedSubject() {
        return FRUGAL_PREFIX + this.subject;
    }
}

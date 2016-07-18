package com.workiva.frugal.transport;

import io.nats.client.*;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

/**
 * FNatsTransport is an extension of FTransport. This is a "stateless" transport
 * in the sense that there is no connection with a server. A request is simply
 * published to a subject and responses are received on another subject. This
 * assumes requests/responses fit within a single NATS message.
 */
public class FNatsTransport extends FBaseTransport {

    private static final int FRAME_BUFFER_SIZE = 64;
    public static final int NATS_MAX_MESSAGE_SIZE = 1024 * 1024;
    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsTransport.class);

    private final Connection conn;
    private final String subject;
    private final String inbox;
    private Subscription sub;

    // TODO: Remove this with 2.0
    private final boolean isTTransport;

    /**
     * Creates a new FTransport which uses the NATS messaging system as the underlying transport.
     * A request is simply published to a subject and responses are received on a randomly generated
     * subject. This requires requests to fit within a single NATS message.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     */
    public FNatsTransport(Connection conn, String subject) {
        this(conn, subject, conn.newInbox());
    }

    /**
     * Creates a new FTransport which uses the NATS messaging system as the underlying transport.
     * A request is simply published to a subject and responses are received on a specified subject.
     * This requires requests to fit within a single NATS message.
     *
     * @param conn    NATS connection
     * @param subject subject to publish requests on
     * @param inbox   subject to receive responses on
     */
    public FNatsTransport(Connection conn, String subject, String inbox) {
        super(NATS_MAX_MESSAGE_SIZE - 4, FRAME_BUFFER_SIZE, LOGGER);
        this.conn = conn;
        this.subject = subject;
        this.inbox = inbox;

        // TODO: Remove this with 2.0
        this.isTTransport = false;
    }

    /**
     * TODO: Remove this with 2.0
     *
     * @deprecated
     */
    FNatsTransport(Connection conn, String subject, String inbox, boolean isTTransport) {
        super(NATS_MAX_MESSAGE_SIZE, FRAME_BUFFER_SIZE, LOGGER);
        this.conn = conn;
        this.subject = subject;
        this.inbox = inbox;
        this.isTTransport = isTTransport;
    }

    @Override
    public synchronized boolean isOpen() {
        return sub != null && conn.getState() == Constants.ConnState.CONNECTED;
    }

    /**
     * Subscribes to the configured inbox subject.
     *
     * @throws TTransportException
     */
    @Override
    public synchronized void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw getClosedConditionException(conn, "open:");
        }
        if (sub != null) {
            throw new TTransportException(TTransportException.ALREADY_OPEN, "NATS transport already open");
        }
        sub = conn.subscribe(inbox, new Handler());
    }

    protected class Handler implements io.nats.client.MessageHandler {
        public void onMessage(Message message) {
            // TODO: Remove this with 2.0
            if (isTTransport) {
                try {
                    frameBuffer.put(message.getData());
                } catch (InterruptedException ignored) {
                }
            } else {
                try {
                    execute(message.getData());
                } catch (TException ignored) {
                }
            }
        }

    }

    /**
     * Unsubscribes from the inbox subject and closes the response buffer.
     */
    @Override
    public synchronized void close() {
        if (sub == null) {
            return;
        }
        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warn("NATS transport could not unsubscribe from subscription: " + e.getMessage());
        }
        sub = null;
        close(null);
    }

    /**
     * Reads up to len bytes into buffer buf, starting at offset off.
     *
     * @throws TTransportException
     */
    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (!isTTransport) {
            throw new TTransportException("Do not call read directly on FTransport");
        }
        if (!isOpen()) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }
        return super.read(bytes, off, len);
    }

    /**
     * Writes the bytes to a buffer. Throws FMessageSizeException if the buffer exceeds 1MB.
     *
     * @throws TTransportException
     */
    @Override
    public void write(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn, "write:");
        }
        super.write(bytes, off, len);
    }

    /**
     * Sends the buffered bytes over NATS.
     *
     * @throws TTransportException
     */
    @Override
    public void flush() throws TTransportException {
        if (!isOpen()) {
            throw getClosedConditionException(conn, "flush:");
        }

        if (!isRequestData()) {
            return;
        }

        // TODO: Remove TTransport check with 2.0
        byte[] data;
        if (isTTransport) {
            data = getRequestBytes();
        } else {
            data = getFramedRequestBytes();
        }

        try {
            conn.publish(subject, inbox, data);
        } catch (IOException e) {
            throw new TTransportException("flush: unable to publish data: " + e.getMessage());
        }
    }

    private static TTransportException getClosedConditionException(Connection conn, String prefix) {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            return new TTransportException(TTransportException.NOT_OPEN,
                    String.format("%s NATS client not connected (has status %s)", prefix, conn.getState().name()));
        }
        return new TTransportException(TTransportException.NOT_OPEN,
                String.format("%s NATS Transport not open", prefix));
    }
}

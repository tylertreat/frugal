package com.workiva.frugal.transport;

import io.nats.client.*;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

import java.io.IOException;
import java.io.PipedInputStream;
import java.io.PipedOutputStream;
import java.nio.ByteBuffer;
import java.util.Timer;
import java.util.TimerTask;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.logging.Logger;

/**
 * TNatsServiceTransport is an extension of thrift.TTransport exclusively used for services which uses NATS as the
 * underlying transport. Message frames are limited to 1MB in size.
 */
public class TNatsServiceTransport extends TTransport {

    // NATS limits messages to 1MB.
    public static final int NATS_MAX_MESSAGE_SIZE = 1024 * 1024;

    protected static final int MAX_MISSED_HEARTBEATS = 3;

    private static final String DISCONNECT = "DISCONNECT";

    private Connection conn;
    private PipedOutputStream writer;
    private PipedInputStream reader;
    private ByteBuffer writeBuffer;
    private Subscription sub;
    private String listenTo;
    private String writeTo;
    private AsyncSubscription heartbeatSub;
    private String heartbeatListen;
    private String heartbeatReply;
    private long heartbeatInterval;
    private Timer heartbeatTimer;
    private AtomicInteger missedHeartbeats;

    private static Logger LOGGER = Logger.getLogger(TNatsServiceTransport.class.getName());

    private TNatsServiceTransport(Connection conn, String listenTo, String writeTo, String heartbeatListen,
                                  String heartbeatReply, long heartbeatInterval) {
        this.conn = conn;
        this.listenTo = listenTo;
        this.writeTo = writeTo;
        this.heartbeatListen = heartbeatListen;
        this.heartbeatReply = heartbeatReply;
        this.heartbeatInterval = heartbeatInterval;
        this.missedHeartbeats = new AtomicInteger(0);
    }

    /**
     * Returns a new thrift TTransport which uses the NATS messaging system as the
     * underlying transport. It performs a handshake with a server listening on the
     * given NATS subject. This TTransport can only be used with FNatsServer.
     */
    public static TNatsServiceTransport client(Connection conn, String subject, long timeout)
            throws TTransportException, TimeoutException {
        Message message;
        try {
            message = conn.request(subject, null, timeout);
        } catch (IOException e) {
            throw new TTransportException(e);
        }
        String reply = message.getReplyTo();
        if (reply == null || reply.isEmpty()) {
            throw new TTransportException("No reply subject on connect.");
        }

        String[] subjects = new String(message.getData()).split(" ");
        if (subjects.length != 3) {
            throw new TTransportException("Invalid connect message.");
        }

        String heartbeatListen = subjects[0];
        String heartbeatReply = subjects[1];
        int deadline;
        try {
            deadline = Integer.parseInt(subjects[2]);
        } catch (NumberFormatException e) {
            throw new TTransportException("Connection deadline not an integer.", e);
        }

        long heartbeatInterval = 0;
        if (deadline > 0) {
            heartbeatInterval = deadline;
        }

        return new TNatsServiceTransport(
                conn, message.getSubject(), reply, heartbeatListen, heartbeatReply, heartbeatInterval
        );
    }

    /**
     * Returns a new thrift TTransport which uses the NATS messaging system as the
     * underlying transport. This TTransport can only be used with FNatsServer.
     */
    public static TNatsServiceTransport server(Connection conn, String listenTo, String writeTo) {
        return new TNatsServiceTransport(conn, listenTo, writeTo, "", "", 0);
    }

    @Override
    public boolean isOpen() {
        return conn.getState() == Constants.ConnState.CONNECTED && sub != null;
    }

    @Override
    public void open() throws TTransportException {
        if (conn.getState() != Constants.ConnState.CONNECTED) {
            throw new TTransportException(TTransportException.NOT_OPEN,
                    "NATS not connected, has status " + conn.getState());
        }

        if (listenTo == null || "".equals(listenTo) || writeTo == null || "".equals(writeTo)) {
            throw new TTransportException("listenTo and writeTo cannot be empty.");
        }

        writeBuffer = ByteBuffer.allocate(NATS_MAX_MESSAGE_SIZE);

        try {
            writer = new PipedOutputStream();
            reader = new PipedInputStream(writer);
        } catch (IOException e) {
            throw new TTransportException(e);
        }

        sub = conn.subscribe(listenTo, new MessageHandler() {
            @Override
            public void onMessage(Message msg) {
                if (DISCONNECT.equals(msg.getReplyTo())) {
                    close();
                    return;
                }
                try {
                    writer.write(msg.getData());
                    writer.flush();
                } catch (IOException e) {
                    LOGGER.warning("could not write incoming data to buffer" + e.getMessage());
                }
            }
        });

        if (heartbeatInterval > 0) {
            startTimer();
            heartbeatSub = conn.subscribe(heartbeatListen, new MessageHandler() {

                @Override
                public void onMessage(Message message) {
                    receiveHeartbeat();
                    conn.publish(heartbeatReply, null);
                }
            });

        }
    }

    private void startTimer() {
        heartbeatTimer = new Timer();
        heartbeatTimer.schedule(new TimerTask() {
            @Override
            public void run() {
                missedHeartbeat();
            }
        }, heartbeatInterval);
    }

    private void missedHeartbeat() {
        int missed = missedHeartbeats.getAndIncrement();
        if (missed >= MAX_MISSED_HEARTBEATS) {
            LOGGER.warning("missed " + missed + " heartbeats from peer, closing transport");
            close();
            return;
        }
        startTimer();
    }

    private void receiveHeartbeat() {
        heartbeatTimer.cancel();
        missedHeartbeats.set(0);
        startTimer();
    }

    @Override
    public void close() {
        if (!isOpen()) {
            return;
        }
        // Signal remote peer for a graceful disconnect.
        conn.publish(writeTo, DISCONNECT, null);

        if (heartbeatSub != null) {
            try {
                heartbeatSub.unsubscribe();
            } catch (IOException e) {
                LOGGER.warning("could not unsubscribe from heartbeat subscription. " + e.getMessage());
            }
            heartbeatSub = null;
        }
        if (heartbeatTimer != null) {
            heartbeatTimer.cancel();
            heartbeatTimer = null;
        }

        try {
            sub.unsubscribe();
        } catch (IOException e) {
            LOGGER.warning("could not unsubscribe from inbox subscription. " + e.getMessage());
        }
        sub = null;

        try {
            writer.close();
        } catch (IOException e) {
            LOGGER.warning("could not close write buffer. " + e.getMessage());
        }
    }

    @Override
    public int read(byte[] bytes, int off, int len) throws TTransportException {
        if (!isOpen()) {
            throw new TTransportException(TTransportException.END_OF_FILE);
        }
        try {
            int bytesRead = this.reader.read(bytes, off, len);
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
        conn.publish(writeTo, data);
        writeBuffer.clear();
    }
}

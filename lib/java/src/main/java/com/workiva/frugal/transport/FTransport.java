package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FAsyncCallback;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import com.workiva.frugal.protocol.FRegistryImpl;
import com.workiva.frugal.transport.monitor.FTransportMonitor;
import com.workiva.frugal.transport.monitor.MonitorRunner;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Arrays;

/**
 * FTransport is comparable to Thrift's TTransport in that it represent the transport
 * layer for frugal clients. However, frugal is callback based and sends only framed data.
 * Therefore, instead of exposing <code>read</code>, <code>write</code>, and <code>flush</code>,
 * the transport has a simple <code>send</code> method that sends framed frugal messages.
 * To handle callback data, an FTransport also has an FRegistry, so it provides methods for
 * registering and unregistering an FAsyncCallback to an FContext.
 */
public abstract class FTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FTransport.class);

    public static final int REQUEST_TOO_LARGE = 100;
    public static final int RESPONSE_TOO_LARGE = 101;

    private volatile FTransportClosedCallback closedCallback;
    private volatile FTransportClosedCallback monitor;
    private boolean isOpen;

    protected FRegistry registry = new FRegistryImpl();
    protected int requestSizeLimit;

    public synchronized boolean isOpen() {
        return isOpen;
    }

    /**
     * Opens the transport.
     *
     * @throws TTransportException
     */
    public synchronized void open() throws TTransportException {
        isOpen = true;
    }

    /**
     * Closes the transport.
     */
    public void close() {
        close(null);
    }

    /**
     * Close registry and signal close.
     *
     * @param cause Exception if not a clean close (null otherwise)
     */
    protected synchronized void close(final Exception cause) {
        registry.close();
        isOpen = false;
        signalClose(cause);
    }

    /**
     * Send the given framed frugal payload over the transport. Implementations of <code>send</code>
     * should be thread-safe.
     *
     * @param payload framed frugal bytes
     * @throws TTransportException
     */
    public abstract void send(byte[] payload) throws TTransportException;

    /**
     * Get the maximum request size permitted by the transport. If <code>getRequestSizeLimit</code>
     * returns a non-positive number, the transport is assumed to have no request size limit.
     *
     * @return the request size limit
     */
    public int getRequestSizeLimit() {
        return requestSizeLimit;
    }

    /**
     * Register a callback for the given FContext.
     *
     * @param context  the FContext to register.
     * @param callback the callback to register.
     */
    public synchronized void register(FContext context, FAsyncCallback callback) throws TException {
        registry.register(context, callback);
    }

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    public synchronized void unregister(FContext context) throws TException {
        registry.unregister(context);
    }

    /**
     * Set the closed callback for the FTransport.
     *
     * @param closedCallback
     */
    public synchronized void setClosedCallback(FTransportClosedCallback closedCallback) {
        this.closedCallback = closedCallback;
    }

    /**
     * Starts a monitor that can watch the health of, and reopen, the transport.
     *
     * @param monitor the FTransportMonitor to set.
     */
    public synchronized void setMonitor(FTransportMonitor monitor) {
        LOGGER.info("FTransport Monitor: Beginning to monitor transport...");
        this.monitor = new MonitorRunner(monitor, this);
    }

    /**
     * Execute a frugal frame (NOTE: this frame must include the frame size).
     *
     * @param frame frugal frame
     * @throws TException
     */
    protected void executeFrame(byte[] frame) throws TException {
        registry.execute(Arrays.copyOfRange(frame, 4, frame.length));
    }

    protected synchronized void signalClose(final Exception cause) {
        if (closedCallback != null) {
            closedCallback.onClose(cause);
        }
        if (monitor != null) {
            new Thread(() -> monitor.onClose(cause), "transport-monitor").start();
        }
    }
}

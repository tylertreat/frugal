package com.workiva.frugal.transport;

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
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.TimeUnit;

/**
 * FTransport is comparable to Thrift's TTransport in that it represent the transport
 * layer for frugal clients. However, frugal is callback based and sends only framed data.
 * Therefore, instead of exposing <code>read</code>, <code>write</code>, and <code>flush</code>,
 * the transport has a simple <code>request</code> method that sends framed frugal messages and
 * returns the response. Before calling <code>request</code> with a given <code>FContext</code>,
 * <code>assignOpId</code> must be called with the context.
 */
public abstract class FTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FTransport.class);

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
     * Assign an opid to the given <code>FContext</code>.
     *
     * @param context <code>FContext</code> to assign an opid.
     * @throws TTransportException if the given context is already registered to a callback.
     */
    public void assignOpId(FContext context) throws TTransportException {
        registry.assignOpId(context);
    }

    /**
     * Send the given framed frugal payload over the transport and returns the response.
     * Implementations of <code>request</code> should be thread-safe.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param oneway indicates to the transport that this is a one-way request. Transport implementations
     *               should return <code>null</code> if <code>oneway</code> is <code>true</code>
     * @param payload framed frugal bytes
     * @return the response bytes
     * @throws TTransportException
     */
    public abstract byte[] request(FContext context, boolean oneway, byte[] payload) throws TTransportException;

    /**
     * Helper method for implementations interacting with the registry.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param oneway indicates to the transport that this is a one-way request. Will return <code>null</code>
     *               if <code>oneway</code> is <code>true</code>
     * @param requestFlusher that flushes the request data.
     * @return the response bytes
     * @throws TTransportException
     */
    protected byte[] request(FContext context, boolean oneway, RequestFlusher requestFlusher)
            throws TTransportException {
        BlockingQueue<byte[]> queue = new ArrayBlockingQueue<>(1);
        if (!oneway) {
            registry.register(context, queue);
        }

        requestFlusher.flush();

        byte[] response;
        try {
            response = queue.poll(context.getTimeout(), TimeUnit.MILLISECONDS);
        } catch (InterruptedException e) {
            throw new TTransportException(TTransportException.TIMED_OUT, "request: timed out");
        } finally {
            registry.unregister(context);
        }

        if (response == FRegistry.POISON_PILL) {
            throw new TTransportException(TTransportException.NOT_OPEN, "request: transport closed, request canceled");
        }

        return response;
    }

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

    /**
     * Helper class for implementations to use when interacting with the registry stored on the
     * transport. Calling flush should flush the request data.
     */
    protected interface RequestFlusher {
        /**
         * Flush the request data.
         *
         * @throws TTransportException
         */
        void flush() throws TTransportException;
    }
}

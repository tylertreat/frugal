package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.protocol.FAsyncCallback;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import com.workiva.frugal.transport.monitor.FTransportMonitor;
import com.workiva.frugal.transport.monitor.MonitorRunner;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * FTransport is Frugal's equivalent of Thrift's TTransport. FTransport extends
 * TTransport and exposes some additional methods. An FTransport typically has an
 * FRegistry, so it provides methods for setting the FRegistry and registering and
 * unregistering an FAsyncCallback to an FContext. It also allows a way for
 * setting an FTransportMonitor and a high-water mark provided by an FServer.
 * <p/>
 * FTransport wraps a TTransport, meaning all existing TTransport implementations
 * will work in Frugal. However, all FTransports must used a framed protocol,
 * typically implemented by wrapping a TFramedTransport.
 * <p/>
 * Most Frugal language libraries include an FMuxTransport implementation, which
 * uses a worker pool to handle messages in parallel.
 */
public abstract class FTransport extends TTransport {

    private static final Logger LOGGER = LoggerFactory.getLogger(FTransport.class);

    public static final int REQUEST_TOO_LARGE = 100;
    public static final int RESPONSE_TOO_LARGE = 101;
    public static final long DEFAULT_WATERMARK = 5 * 1000;

    private volatile FClosedCallback fClosedCallback;
    private volatile FTransportClosedCallback closedCallback;
    private volatile FTransportClosedCallback monitor;
    protected long highWatermark = DEFAULT_WATERMARK;
    protected FRegistry registry;
    private boolean isOpen;

    @Override
    public synchronized boolean isOpen() {
        return isOpen;
    }

    @Override
    public synchronized void open() throws TTransportException {
        isOpen = true;
    }

    @Override
    public synchronized void close() {
        isOpen = false;
    }

    /**
     * Set the FRegistry on the FTransport.
     *
     * @param registry FRegistry to set on the FTransport.
     */
    public synchronized void setRegistry(FRegistry registry) {
        if (registry == null) {
            throw new RuntimeException("registry cannot by null");
        }
        if (this.registry != null) {
            return;
        }
        this.registry = registry;
    }

    /**
     * Register a callback for the given FContext.
     *
     * @param context  the FContext to register.
     * @param callback the callback to register.
     */
    public synchronized void register(FContext context, FAsyncCallback callback) throws TException {
        if (registry == null) {
            throw new FException("registry not set");
        }
        registry.register(context, callback);
    }

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    public synchronized void unregister(FContext context) throws TException {
        if (registry == null) {
            throw new FException("registry not set");
        }
        registry.unregister(context);
    }

    protected synchronized FRegistry getRegistry() {
        return registry;
    }

    /**
     * Set the closed callback for the FTransport.
     *
     * @param closedCallback
     * @deprecated use {@link #setClosedCallback(FTransportClosedCallback)} instead.
     */
    @Deprecated
    public synchronized void setClosedCallback(FClosedCallback closedCallback) {
        this.fClosedCallback = closedCallback;
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
     * Sets the maximum amount of time a frame is allowed to await processing
     * before triggering transport overload logic.
     *
     * @param watermark the watermark time in milliseconds.
     */
    public synchronized void setHighWatermark(long watermark) {
        this.highWatermark = watermark;
    }

    @Override
    public int read(byte[] buff, int off, int len) throws TTransportException {
        throw new RuntimeException("Do not call read directly on FTransport");
    }

    protected synchronized long getHighWatermark() {
        return highWatermark;
    }

    protected synchronized void signalClose(final Exception cause) {
        // TODO: Remove deprecated callback in future release.
        if (fClosedCallback != null) {
            fClosedCallback.onClose();
        }
        if (closedCallback != null) {
            closedCallback.onClose(cause);
        }
        if (monitor != null) {
            new Thread(() -> {
                monitor.onClose(cause);
            }, "transport-monitor").start();
        }
    }

}

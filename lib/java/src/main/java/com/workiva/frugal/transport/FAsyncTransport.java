package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.FException;
import com.workiva.frugal.protocol.HeaderUtils;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.util.Map;
import java.util.Objects;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantLock;

/**
 * FAsyncTransport is an extension of FTransport that asynchronous frameworks can implement.
 * Implementations need only implement <code>flush</code> to send request data and call
 * <code>handleResponse</code> when asynchronous responses are received.
 */
public abstract class FAsyncTransport extends FTransport {

    protected static final String OPID_HEADER = "_opid";
    protected static final byte[] POISON_PILL = new byte[0];

    private final ReentrantLock lock = new ReentrantLock();
    protected Map<Long, BlockingQueue<byte[]>> queueMap = new ConcurrentHashMap<>();

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
        lock.lock();
        queueMap.values().parallelStream()
                .filter(Objects::nonNull)
                .forEach(queue -> queue.offer(POISON_PILL));
        queueMap.clear();
        lock.unlock();
        super.close(cause);
    }

    /**
     * Send the given framed frugal payload over the transport and returns the response.
     *
     * @param context FContext associated with the request (used for timeout and logging)
     * @param oneway indicates to the transport that this is a one-way request. Will return <code>null</code>
     *               if <code>oneway</code> is <code>true</code>
     * @param payload framed frugal bytes
     * @return the response bytes
     * @throws TTransportException
     */
    public byte[] request(FContext context, boolean oneway, byte[] payload) throws TTransportException {
        if (oneway) {
            flush(payload);
            return null;
        }

        BlockingQueue<byte[]> queue = new ArrayBlockingQueue<>(1);
        lock.lock();
        if (queueMap.containsKey(getOpId(context))) {
            lock.unlock();
            throw new TTransportException("request already in flight for context");
        }
        queueMap.put(getOpId(context), queue);
        lock.unlock();

        try {
            flush(payload);

            byte[] response;
            try {
                response = queue.poll(context.getTimeout(), TimeUnit.MILLISECONDS);
            } catch (InterruptedException e) {
                throw new TTransportException("request: interrupted");
            }

            if (response == null) {
                throw new TTransportException(TTransportException.TIMED_OUT, "request: timed out");
            }

            if (response == POISON_PILL) {
                throw new TTransportException(TTransportException.NOT_OPEN,
                        "request: transport closed, request canceled");
            }

            return response;
        } finally {
            lock.lock();
            queueMap.remove(getOpId(context));
            lock.unlock();
        }
    }

    /**
     * Flush the payload to the server. Implementations must not block and must be thread-safe.
     *
     * @param payload framed frugal bytes
     * @throws TTransportException
     */
    protected abstract void flush(byte[] payload) throws TTransportException;

    /**
     * Handles a frugal frame response (NOTE: this frame must NOT include the frame size).
     * Implementations should call this when asynchronous responses are recieved from the server.
     *
     * @param frame frugal frame
     * @throws TException
     */
    protected void handleResponse(byte[] frame) throws TException {
        Map<String, String> headers;
        headers = HeaderUtils.decodeFromFrame(frame);

        long opId;
        try {
            opId = Long.parseLong(headers.get("_opid"));
        } catch (NumberFormatException e) {
            throw new FException("invalid protocol frame: op id not a uint64", e);
        }

        lock.lock();
        BlockingQueue<byte[]> queue = queueMap.get(opId);
        lock.unlock();

        // Ignore unregistered frames
        if (queue == null) {
            return;
        }

        try {
            queue.put(frame);
        } catch (InterruptedException e) {
            throw new TException(e);
        }
    }

    /**
     * Returns the operation id for the FContext. This is a unique long per context. This is protected as operation
     * ids are an internal implementation detail.
     *
     * @return operation id
     */
    protected static long getOpId(FContext context) {
        String opIdStr = context.getRequestHeader(OPID_HEADER);
        if (opIdStr == null) {
            return 0;
        }
        return Long.valueOf(opIdStr);
    }
}

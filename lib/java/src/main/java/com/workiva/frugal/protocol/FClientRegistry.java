package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import com.workiva.frugal.internal.Headers;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;


/**
 * FClientRegistry is intended for use only by Frugal clients.
 * This is only to be used by generated code.
 */
public class FClientRegistry implements FRegistry {

    private static final Logger LOGGER = LoggerFactory.getLogger(FClientRegistry.class);
    private static final AtomicLong NEXT_OP_ID = new AtomicLong(0);

    protected Map<Long, FAsyncCallback> handlers;

    public FClientRegistry() {
        handlers = new ConcurrentHashMap<>();
    }

    /**
     * Register a callback for the given FContext.
     *
     * @param context  the FContext to register.
     * @param callback the callback to register.
     */
    public void register(FContext context, FAsyncCallback callback) throws TException {
        // Assign an opId if one does not exist
        if (context.getOpId() == 0) {
            context.setOpId(NEXT_OP_ID.incrementAndGet());
        }

        if (handlers.containsKey(context.getOpId())) {
            throw new FException("context already registered");
        }
        handlers.put(context.getOpId(), callback);
    }

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    public void unregister(FContext context) {
        if (context == null) {
            return;
        }
        handlers.remove(context.getOpId());
    }

    /**
     * Dispatch a single Frugal message frame.
     *
     * @param frame an entire Frugal message frame.
     */
    public void execute(byte[] frame) throws TException {
        Map<String, String> headers;
        headers = Headers.decodeFromFrame(frame);

        long opId;
        try {
            opId = Long.parseLong(headers.get(FContext.OP_ID));
        } catch (NumberFormatException e) {
            throw new FException("frame missing opId");
        }

        FAsyncCallback callback = handlers.get(opId);
        if (callback == null) {
            LOGGER.info("Got a message for an unregistered context. Dropping.");
            return;
        }
        callback.onMessage(new TMemoryInputTransport(frame));
    }

    /**
     * Interrupt any registered contexts.
     */
    public void close() {
        handlers.clear();
    }
}

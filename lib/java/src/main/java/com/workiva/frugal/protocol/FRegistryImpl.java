package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.util.Map;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;


/**
 * FRegistryImpl is intended for use only by Frugal clients.
 */
public class FRegistryImpl implements FRegistry {

    private static final AtomicLong NEXT_OP_ID = new AtomicLong(0);
    protected static final BlockingQueue NULL_QUEUE = new ArrayBlockingQueue(1);

    protected Map<Long, BlockingQueue<byte[]>> handlers;

    public FRegistryImpl() {
        handlers = new ConcurrentHashMap<>();
    }

    @Override
    public void assignOpId(FContext context) throws TTransportException {
        // An FContext can be reused for multiple requests. Because of this, every
        // time an FContext is registered, it must be assigned a new op id to
        // ensure we can properly correlate responses. We use a monotonically
        // increasing atomic long for this purpose. If the FContext already has
        // an op id, it has been used for a request. We check the handlers map to
        // ensure that request is not still in-flight.
        if (handlers.containsKey(context.getOpId())) {
            throw new TTransportException("context already registered");
        }
        context.setOpId(NEXT_OP_ID.incrementAndGet());
        // Add a placeholder for this opid
        handlers.put(context.getOpId(), NULL_QUEUE);
    }

    @Override
    public void register(FContext context, BlockingQueue<byte[]> queue) {
        handlers.put(context.getOpId(), queue);
    }

    @Override
    public void unregister(FContext context) {
        if (context == null) {
            return;
        }
        handlers.remove(context.getOpId());
    }

    @Override
    public void execute(byte[] frame) throws TException {
        Map<String, String> headers;
        headers = HeaderUtils.decodeFromFrame(frame);

        long opId;
        try {
            opId = Long.parseLong(headers.get(FContext.OPID_HEADER));
        } catch (NumberFormatException e) {
            throw new FException("invalid protocol frame: op id not a uint64", e);
        }

        BlockingQueue<byte[]> queue = handlers.get(opId);
        if (queue == null) {
            return;
        }

        try {
            queue.put(frame);
        } catch (InterruptedException e) {
            throw new TException(e);
        }
    }

    @Override
    public void close() {
        handlers.values().parallelStream()
                .filter(queue -> queue != NULL_QUEUE)
                .forEach(queue -> queue.offer(POISON_PILL));
        handlers.clear();
    }
}

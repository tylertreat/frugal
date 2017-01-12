package com.workiva.frugal.protocol;

import com.workiva.frugal.exception.FException;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.util.Map;
import java.util.Objects;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ConcurrentHashMap;


/**
 * FRegistryImpl is intended for use only by Frugal clients.
 */
public class FRegistryImpl implements FRegistry {

    protected Map<Long, BlockingQueue<byte[]>> queueMap;

    public FRegistryImpl() {
        queueMap = new ConcurrentHashMap<>();
    }

    @Override
    public void register(FContext context, BlockingQueue<byte[]> queue) throws TTransportException {
        if (queueMap.containsKey(context.getOpId())) {
            throw new TTransportException("request already in flight for context");
        }
        queueMap.put(context.getOpId(), queue);
    }

    @Override
    public void unregister(FContext context) {
        if (context == null) {
            return;
        }
        queueMap.remove(context.getOpId());
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

        BlockingQueue<byte[]> queue = queueMap.get(opId);

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

    @Override
    public void close() {
        queueMap.values().parallelStream()
                .filter(Objects::nonNull)
                .forEach(queue -> queue.offer(POISON_PILL));
        queueMap.clear();
    }
}

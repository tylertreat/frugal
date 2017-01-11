package com.workiva.frugal.protocol;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.util.concurrent.BlockingQueue;


/**
 * FRegistry is responsible for multiplexing and handling messages received
 * from the server. An FRegistry is used by an FTransport.
 * <p>
 * When a request is made, an FAsyncCallback is registered to an FContext. When a
 * response for the FContext is received, the FAsyncCallback is looked up,
 * executed, and unregistered.
 */
public interface FRegistry extends AutoCloseable {

    /**
     * Poison pill placed in all registered queues when <code>close</code> is called.
     */
    byte[] POISON_PILL = new byte[0];

    /**
     * Assign an opid to the given <code>FContext</code> and make a placeholder for the given
     * opid in the registry.
     *
     * @param context <code>FContext</code> to assign an opid.
     * @throws TTransportException if the given context is already registered to a callback.
     */
    void assignOpId(FContext context) throws TTransportException;

    /**
     * Register a queue for the given FContext.
     *
     * @param context  the FContext to register.
     * @param queue    the queue to place responses directed at this context.
     */
    void register(FContext context, BlockingQueue<byte[]> queue);

    /**
     * Unregister the callback for the given FContext.
     *
     * @param context the FContext to unregister.
     */
    void unregister(FContext context);

    /**
     * Dispatch a single Frugal message frame.
     *
     * @param frame an entire Frugal message frame.
     * @throws TException if execution failed.
     */
    void execute(byte[] frame) throws TException;

    /**
     * Interrupt any registered contexts.
     */
    void close();
}

package com.workiva.frugal.protocol;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;

import java.util.concurrent.BlockingQueue;


/**
 * FRegistry is responsible for multiplexing and handling messages received
 * from the server. An FRegistry is used by an FTransport.
 * <p>
 * When a request is made, an BlockingQueue is registered to an FContext. When a
 * response for the FContext is received, the queue is looked up and the response
 * is placed in it.
 */
public interface FRegistry extends AutoCloseable {

    /**
     * Poison pill placed in all registered queues when <code>close</code> is called.
     */
    byte[] POISON_PILL = new byte[0];

    /**
     * Register a queue for the given FContext.
     *
     * @param context  the FContext to register.
     * @param queue    the queue to place responses directed at this context.
     *
     * @throws TTransportException if the given context is already registered to a queue.
     */
    void register(FContext context, BlockingQueue<byte[]> queue) throws TTransportException;

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

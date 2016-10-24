package com.workiva.frugal.protocol;

import org.apache.thrift.TException;


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
     * Register a callback for the given FContext.
     *
     * @param context  the FContext to register.
     * @param callback the callback to register.
     * @throws TException if the given context is already registered to a callback.
     */
    void register(FContext context, FAsyncCallback callback) throws TException;

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

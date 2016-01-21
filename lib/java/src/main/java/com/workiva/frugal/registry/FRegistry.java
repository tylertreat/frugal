package com.workiva.frugal.registry;

import com.workiva.frugal.FContext;
import org.apache.thrift.TException;

/**
 * Registry is responsible for multiplexing received messages to the appropriate callback.
 */
public interface FRegistry {

    /**
     * Register a callback for the given FContext.
     *
     * @param context the FContext to register.
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

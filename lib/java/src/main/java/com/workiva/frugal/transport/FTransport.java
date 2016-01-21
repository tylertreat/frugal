package com.workiva.frugal.transport;

import com.workiva.frugal.FContext;
import com.workiva.frugal.FException;
import com.workiva.frugal.registry.FAsyncCallback;
import com.workiva.frugal.registry.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;

/**
 * FTransport is a Thrift TTransport for services.
 */
public abstract class FTransport extends TTransport {

    protected volatile FClosedCallback closedCallback;
    protected FRegistry registry;

    /**
     * Set the FRegistry on the FTransport.
     *
     * @param registry FRegistry to set on the FTransport.
     */
    public abstract void setRegistry(FRegistry registry);

    /**
     * Register a callback for the given FContext.
     *
     * @param context the FContext to register.
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

    /**
     * Set the closed callback for the FTransport.
     *
     * @param closedCallback
     */
    public void setClosedCallback(FClosedCallback closedCallback) {
        this.closedCallback = closedCallback;
    }
}

package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FAsyncCallback;
import org.apache.thrift.TException;

/**
 * FSubscriberTransport is used exclusively for scope publishers.
 */
public abstract class FSubscriberTransport {

    /**
     * Opens the Transport to receive messages on the subscription.
     *
     * @param topic the pub/sub topic to subscribe to.
     * @throws TException
     */
    public abstract void subscribe(String topic, FAsyncCallback callback) throws TException;

    /**
     * Closes the transport by unsubscribing from the set topic.
     */
    public abstract void unsubscribe();

}

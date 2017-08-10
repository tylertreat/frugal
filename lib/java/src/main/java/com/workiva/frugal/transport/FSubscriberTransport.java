package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FAsyncCallback;
import org.apache.thrift.TException;

/**
 * FSubscriberTransport is used exclusively for scope publishers.
 */
public interface FSubscriberTransport {

    /**
     * Queries whether the transport is subscribed to a topic.
     *
     * @return True if the transport is subscribed to a topic.
     */
    boolean isSubscribed();

    /**
     * Opens the Transport to receive messages on the subscription.
     *
     * @param topic the pub/sub topic to subscribe to.
     * @throws TException if there was a problem subscribing.
     */
    void subscribe(String topic, FAsyncCallback callback) throws TException;

    /**
     * Closes the transport by unsubscribing from the set topic.
     */
    void unsubscribe();

    /**
     * Remove unsubscribes and removes durably stored information on the broker, if applicable.
     */
    default void remove() throws TException {
        unsubscribe();
    }
}

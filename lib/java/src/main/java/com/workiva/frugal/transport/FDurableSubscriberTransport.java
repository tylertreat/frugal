package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FDurableAsyncCallback;
import org.apache.thrift.TException;

/**
 * FDurableSubscriberTransport is used exclusively for pub/sub scopes. Subscribers
 * use it to durably subscribe to a pub/sub topic.
 */
public interface FDurableSubscriberTransport {

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
    void subscribe(String topic, FDurableAsyncCallback callback) throws TException;

    /**
     * Closes the transport by unsubscribing from the set topic.
     */
    void unsubscribe();

}

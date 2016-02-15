package com.workiva.frugal.transport;

import com.workiva.frugal.transport.FScopeTransport;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

/**
 * FSubscription to a pub/sub topic.
 */
public class FSubscription {

    private String topic;
    private FScopeTransport transport;
    private BlockingQueue<Exception> onError;

    /**
     * Construct a new subscription. This is used only by generated
     * code and should not be called directly.
     *
     * @param topic for the subscription.
     * @param transport for the subscription.
     */
    public FSubscription(String topic, FScopeTransport transport) {
        this.topic = topic;
        this.transport = transport;
        this.onError = new LinkedBlockingQueue<>(1);
    }

    /**
     * Get the subscription topic.
     *
     * @return subscription topic.
     */
    public String getTopic() { return topic; }

    /**
     * Unsubscribe from the topic.
     */
    public void unsubscribe() { transport.close(); }

    /**
     * Queue which returns any error might have occured during
     * the subscription's lifetime. An error indicates that this
     * subscription has been closed.
     *
     * @return The Exception queue.
     */
    public BlockingQueue<Exception> onError() {
        return onError;
    }

    /**
     * Used to indicate an error on the subscription. This is used
     * only by generated code and should not be called directly.
     *
     * @param ex Exception causing the interruption.
     */
    public void signal(Exception ex) {
        try {
            onError.put(ex);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }
}

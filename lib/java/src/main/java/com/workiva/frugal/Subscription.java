package com.workiva.frugal;

import org.apache.thrift.transport.TTransportException;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;

/**
 * Subscription to a pub/sub topic.
 */
public class Subscription {

    private String topic;
    private Transport transport;
    private BlockingQueue<Exception> onError;

    public Subscription(String topic, Transport transport) {
        this.topic = topic;
        this.transport = transport;
        this.onError = new LinkedBlockingQueue<>(1);
    }

    public void unsubscribe() throws TTransportException {
        transport.unsubscribe();
    }

    public BlockingQueue<Exception> onError() {
        return onError;
    }

    public void signal(Exception ex) {
        try {
            onError.put(ex);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }
}

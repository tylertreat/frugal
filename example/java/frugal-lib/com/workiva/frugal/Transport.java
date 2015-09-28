package com.workiva.frugal;

import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.apache.thrift.transport.TTransportFactory;

/**
 * Transport wraps a Thrift TTransport which supports pub/sub.
 */
public interface Transport {

    /**
     * Opens the Transport to receive messages on the subscription.
     *
     * @param topic the pub/sub topic to subscribe to.
     * @throws TTransportException
     */
    void subscribe(String topic) throws TTransportException;

    /**
     * Closes the Transport to stop receiving messages on the subscription.
     *
     * @throws TTransportException
     */
    void unsubscribe() throws TTransportException;

    /**
     * Prepares the Transport for publishing to the given topic.
     *
     * @param topic the pub/sub topic to publish on.
     */
    void preparePublish(String topic);

    /**
     * Returns the wrapped Thrift TTransport.
     *
     * @return wrapped transport.
     */
    TTransport thriftTransport();

    /**
     * Wraps the underlying TTransport with the TTransport returned by the given
     * TTransportFactory.
     *
     * @param proxy the factory to wrap the transport with.
     */
    void applyProxy(TTransportFactory proxy);

}

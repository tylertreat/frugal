package com.workiva.frugal.transport;

import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;

/**
 * FPublisherTransport extends Thrift's TTransport and is used exclusively
 * for scope publishers.
 */
public abstract class FPublisherTransport extends TTransport {
    /**
     * FPublsiherTransports do not support read.
     * This will throw an UnsupportedOperationException when called.
     */
    public int read(byte[] buf, int off, int len) throws TTransportException {
        throw new UnsupportedOperationException("FPublisherTransport does not support read.");
    }

    /**
     * Sets the publish topic and locks the transport for exclusive access.
     *
     * @param topic the pub/sub topic to publish on.
     * @throws TException
     */
    public abstract void lockTopic(String topic) throws TException;

    /**
     * Unsets the publish topic and unlocks the transport.
     *
     * @throws TException
     */
    public abstract void unlockTopic() throws TException;
}

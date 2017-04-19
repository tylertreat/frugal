package com.workiva.frugal.transport;

import org.apache.thrift.transport.TTransportException;

/**
 * FDurablePublisherTransport is used exclusively for pub/sub scopes. Publishers
 * use it to publish messages durably to a topic.
 */
public interface FDurablePublisherTransport {

    /**
     * Queries whether the transport is open.
     *
     * @return True if the transport is open.
     */
    boolean isOpen();

    /**
     * Opens the transport.
     *
     * @throws TTransportException if the transport could not be opened.
     */
    void open() throws TTransportException;

    /**
     * Closes the transport.
     */
    void close();

    /**
     * Get the maximum publish size permitted by the transport. If
     * {@link #getPublishSizeLimit} returns a non-positive number, the
     * transport is assumed to have no publish size limit.
     *
     * @return the publish size limit
     */
    int getPublishSizeLimit();

    /**
     * Publish the given framed frugal payload over the transport. Implementations of <code>publish</code>
     * should be thread-safe.
     *
     * @param topic the topic on which to publish the payload
     * @param groupId the group id associated with the durable queue
     * @param payload framed frugal bytes
     * @throws TTransportException if publishing the payload failed
     */
    void publish(String topic, String groupId, byte[] payload) throws
    TTransportException;
}

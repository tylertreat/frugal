package com.workiva.frugal.transport;

/**
 * FPublisherTransportFactory produces FPublisherTransports and is typically
 * used by an FScopeProvider.
 */
public interface FPublisherTransportFactory {

    /**
     * Get a new FPublisherTransport instance.
     *
     * @return A new FPublisherTransport instance.
     */
    FPublisherTransport getTransport();
}

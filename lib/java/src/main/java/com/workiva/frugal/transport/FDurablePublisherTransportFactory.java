package com.workiva.frugal.transport;

/**
 * FDurablePublisherTransportFactory produces FPublisherTransports and is typically
 * used by an FScopeProvider.
 */
public interface FDurablePublisherTransportFactory {

    /**
     * Get a new FPublisherTransport instance.
     *
     * @return A new FDurablePublisherTransport instance.
     */
    FDurablePublisherTransport getTransport();
}

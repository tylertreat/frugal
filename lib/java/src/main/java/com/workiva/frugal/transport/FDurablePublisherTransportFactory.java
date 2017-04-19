package com.workiva.frugal.transport;

/**
 * FDurablePublisherTransportFactory produces FPublisherTransports and
 * is typically used by an FDurableScopeProvider.
 */
public interface FDurablePublisherTransportFactory {

    /**
     * Get a new FDurablePublisherTransport instance.
     *
     * @return A new FDurablePublisherTransport instance.
     */
    FDurablePublisherTransport getTransport();
}

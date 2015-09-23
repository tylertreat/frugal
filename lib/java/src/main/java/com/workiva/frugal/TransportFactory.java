package com.workiva.frugal;

/**
 * TransportFactory is responsible for creating new Frugal Transports.
 */
public interface TransportFactory {

    /**
     * Creates a new Transport.
     *
     * @return new Frugal Transport for pub/sub.
     */
    Transport getTransport();

}

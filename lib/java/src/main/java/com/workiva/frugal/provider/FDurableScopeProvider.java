package com.workiva.frugal.provider;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.FDurablePublisherTransport;
import com.workiva.frugal.transport.FDurablePublisherTransportFactory;
import com.workiva.frugal.transport.FDurableSubscriberTransport;
import com.workiva.frugal.transport.FDurableSubscriberTransportFactory;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * FDurableScopeProvider produces FDurablePublisherTransports, FDurableSubscriberTransports, and
 * FProtocols for use by pub/sub scopes. It does this by wrapping an
 * FDurablePublisherTransportFactory, an FDurableSubscriberTransportFactory, and an
 * FProtocolFactory. This also provides a shim for adding middleware to a
 * publisher or subscriber.
 */
public class FDurableScopeProvider {

    /**
     * Publisher of this scope.
     */
    public static class Publisher {
        private FDurablePublisherTransport transport;
        private FProtocolFactory protocolFactory;

        private Publisher(FDurablePublisherTransport t, FProtocolFactory pf) {
            transport = t;
            protocolFactory = pf;
        }

        public FDurablePublisherTransport getTransport() {
            return transport;
        }

        public FProtocolFactory getProtocolFactory() {
            return protocolFactory;
        }
    }

    /**
     * Subscriber of this scope.
     */
    public static class Subscriber {
        private FDurableSubscriberTransport transport;
        private FProtocolFactory protocolFactory;

        private Subscriber(FDurableSubscriberTransport t, FProtocolFactory pf) {
            transport = t;
            protocolFactory = pf;
        }

        public FDurableSubscriberTransport getTransport() {
            return transport;
        }

        public FProtocolFactory getProtocolFactory() {
            return protocolFactory;
        }
    }

    private FDurablePublisherTransportFactory publisherTransportFactory;
    private FDurableSubscriberTransportFactory subscriberTransportFactory;
    private FProtocolFactory protocolFactory;
    private List<ServiceMiddleware> middleware;

    public FDurableScopeProvider(FDurablePublisherTransportFactory ptf, FDurableSubscriberTransportFactory stf,
                          FProtocolFactory pf, ServiceMiddleware ...middleware) {
        publisherTransportFactory = ptf;
        subscriberTransportFactory = stf;
        protocolFactory = pf;
        this.middleware = Arrays.asList(middleware);
    }

    /**
     * Returns a new Publisher containing an FPublisherTransport and FProtocolFactory
     * used for publishing.
     *
     * @return Publisher with FPublisherTransport and FProtocol.
     */
    public Publisher buildPublisher() {
        FDurablePublisherTransport transport = publisherTransportFactory.getTransport();
        return new Publisher(transport, protocolFactory);
    }

    /**
     * Returns a new Subscriber containing an FSubscriberTransport and FProtocolFactory
     * used for subscribing.
     *
     * @return SubscriberClient with FSubscriberTransport and FProtocol.
     */
    public Subscriber buildSubscriber() {
        FDurableSubscriberTransport transport = subscriberTransportFactory.getTransport();
        return new Subscriber(transport, protocolFactory);
    }

    public List<ServiceMiddleware> getMiddleware() {
        return new ArrayList<>(middleware);
    }
}

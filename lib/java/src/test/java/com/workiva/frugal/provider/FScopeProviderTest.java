package com.workiva.frugal.provider;

import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.*;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FScopeProvider}.
 */
@RunWith(JUnit4.class)
public class FScopeProviderTest {

    @Test
    public void testProvide() throws Exception {
        FPublisherTransportFactory publisherTransportFactory = mock(FPublisherTransportFactory.class);
        FSubscriberTransportFactory subscriberTransportFactory = mock(FSubscriberTransportFactory.class);
        FProtocolFactory protocolFactory = mock(FProtocolFactory.class);

        FScopeProvider provider = new FScopeProvider(
                publisherTransportFactory,
                subscriberTransportFactory,
                protocolFactory
        );

        // Validate buildPublisher works as intended
        FPublisherTransport publisherTransport = mock(FPublisherTransport.class);
        FProtocol fProtocol = mock(FProtocol.class);

        when(publisherTransportFactory.getTransport()).thenReturn(publisherTransport);
        when(protocolFactory.getProtocol(publisherTransport)).thenReturn(fProtocol);

        FScopeProvider.PublisherClient publisherClient = provider.buildPublisher();

        assertEquals(publisherTransport, publisherClient.getTransport());
        assertEquals(fProtocol, publisherClient.getProtocol());

        // Validate buildSubscriber works as intended
        FSubscriberTransport subscriberTransport = mock(FSubscriberTransport.class);
        when(subscriberTransportFactory.getTransport()).thenReturn(subscriberTransport);

        FScopeProvider.SubscriberClient subscriberClient = provider.buildSubscriber();

        assertEquals(subscriberTransport, subscriberClient.getTransport());
        assertEquals(protocolFactory, subscriberClient.getProtocolFactory());


    }
}

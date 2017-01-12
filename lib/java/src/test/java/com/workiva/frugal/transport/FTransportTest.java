package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.util.concurrent.BlockingQueue;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotNull;
import static org.junit.Assert.assertNull;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;


/**
 * Tests for {@link FTransport}.
 */
@RunWith(JUnit4.class)
public class FTransportTest {

    private FTransport transport;

    @Before
    public void setUp() throws Exception {
        transport = new FTransportTester();
    }

    /**
     * Ensures request registers context, calls RequestFlusher, returns response, and finally unregisters context.
     */
    @Test
    public void testRequest() throws TTransportException {
        byte[] expectedResponse = "hi".getBytes();
        MockRegistry mockRegistry = new MockRegistry(expectedResponse);
        transport.registry = mockRegistry;

        FContext context = new FContext();
        FTransport.RequestFlusher requestFlusher = mock(FTransport.RequestFlusher.class);
        byte[] actualResponse = transport.request(context, false, requestFlusher);

        assertArrayEquals(expectedResponse, actualResponse);
        assertEquals(context, mockRegistry.registeredContext);
        assertNotNull(mockRegistry.registeredQueue);
        assertEquals(context, mockRegistry.unregisteredContext);

        verify(requestFlusher, times(1)).flush();
    }

    /**
     * Ensures oneway request calls RequestFlusher and returns null.
     */
    @Test
    public void testRequestOneway() throws TTransportException {
        byte[] expectedResponse = "hi".getBytes();
        MockRegistry mockRegistry = new MockRegistry(expectedResponse);
        transport.registry = mockRegistry;

        FContext context = new FContext();
        FTransport.RequestFlusher requestFlusher = mock(FTransport.RequestFlusher.class);
        byte[] actualResponse = transport.request(context, true, requestFlusher);

        assertNull(actualResponse);
        assertNull(mockRegistry.registeredContext);
        assertNull(mockRegistry.registeredQueue);
        assertNull(mockRegistry.unregisteredContext);

        verify(requestFlusher, times(1)).flush();
    }

    /**
     * Ensures request timeout throws TTransportException.
     */
    @Test(expected = TTransportException.class)
    public void testRequestTimeout() throws TTransportException {
        transport.registry = mock(FRegistry.class);
        FContext context = new FContext();
        context.setTimeout(10);
        transport.request(context, false, () -> {});
    }

    /**
     * Ensures TTransportException is thrown if poison pill placed in registered queue.
     */
    @Test(expected = TTransportException.class)
    public void testRequestPoisonPill() throws TTransportException {
        transport.registry = new MockRegistry(FRegistry.POISON_PILL);
        transport.request(new FContext(), false, () -> {});
    }

    class FTransportTester extends FTransport {


        @Override
        public boolean isOpen() {
            return false;
        }

        @Override
        public void open() throws TTransportException {
        }

        @Override
        public void close() {
        }

        @Override
        public byte[] request(FContext context, boolean oneway, byte[] payload) throws TTransportException {
            return new byte[0];
        }

    }

    static class MockRegistry implements FRegistry {

        FContext registeredContext;
        FContext unregisteredContext;
        BlockingQueue registeredQueue;
        byte[] response;

        MockRegistry(byte[] response) {
            this.response = response;
        }

        @Override
        public void register(FContext context, BlockingQueue<byte[]> queue) {
            registeredContext = context;
            registeredQueue = queue;
            try {
                queue.put(response);
            } catch (Exception ignored) {
            }
        }

        @Override
        public void unregister(FContext context) {
            unregisteredContext = context;
        }

        @Override
        public void execute(byte[] frame) throws TException {

        }

        @Override
        public void close() {

        }
    }
}

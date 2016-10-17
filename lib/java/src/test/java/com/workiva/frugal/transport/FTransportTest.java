package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FAsyncCallback;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

/**
 * Tests for {@link FTransport}.
 */
@RunWith(JUnit4.class)
public class FTransportTest {

    private final FContext context = new FContext();
    private FAsyncCallback callback = new FAsyncCallback() {
        @Override
        public void onMessage(TTransport transport) throws TException {

        }
    };
    private FTransport transport;

    @Before
    public void setUp() throws Exception {
        transport = new FTransportTester();
    }

    @Test
    public void testRegister() throws Exception {
        FRegistry registry = mock(FRegistry.class);
        transport.registry = registry;
        transport.register(context, callback);
        verify(registry).register(context, callback);
    }

    @Test
    public void testUnregister() throws Exception {
        FRegistry registry = mock(FRegistry.class);
        transport.registry = registry;
        transport.unregister(context);
        verify(registry).unregister(context);
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
        public void send(byte[] payload) throws TTransportException {
        }
    }
}

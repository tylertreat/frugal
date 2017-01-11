package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.Mockito;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;

import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertTrue;
import static org.junit.Assert.fail;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for FAdapterTransport.
 */
public class FAdapterTransportTest {

    private TTransport mockTr;
    private FAdapterTransport tr;

    @Before
    public void setUp() {
        mockTr = mock(TTransport.class);
        tr = new FAdapterTransport(mockTr);
    }

    /**
     * Ensures open throws a TTransportException if the wrapped transport fails to open.
     */
    @Test(expected = TTransportException.class)
    public void testOpen_error() throws TTransportException {
        doThrow(TTransportException.class).when(mockTr).open();

        tr.open();

        verify(mockTr).open();
    }

    /**
     * Ensures open throws an ALREADY_OPEN TTransportException if the transport is already open.
     */
    @Test
    public void testOpen_alreadyOpen() throws TTransportException {
        when(mockTr.isOpen()).thenReturn(true);

        tr.open();

        try {
            tr.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportException.ALREADY_OPEN, e.getType());
        }

        verify(mockTr).open();
        verify(mockTr).isOpen();
    }

    /**
     * Ensures open opens the underlying transport and starts the read thread and close shuts down the executor.
     */
    @Test
    public void testOpenClose() throws TTransportException {
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        tr.setExecutorFactory(mockExecutorFactory);
        FRegistry mockRegistry = mock(FRegistry.class);
        tr.registry = mockRegistry;
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        when(mockTr.isOpen()).thenReturn(true);

        tr.open();

        verify(mockExecutorFactory).newExecutor();
        verify(mockExecutor).submit(any(Runnable.class));

        assertTrue(tr.isOpen());

        tr.close();

        verify(mockRegistry).close();
        verify(mockTr).close();
        verify(mockExecutor).shutdownNow();

        assertFalse(tr.isOpen());
    }

    /**
     * Ensures TransportReader reads frames from the transport and executes them on the registry. When it encounters an
     * EOF it closes the transport and returns.
     */
    @Test
    public void testTransportReader() throws TException {
        FRegistry mockRegistry = mock(FRegistry.class);
        tr.registry = mockRegistry;
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        tr.setExecutorFactory(mockExecutorFactory);
        tr.open();
        when(mockTr.isOpen()).thenReturn(true);
        when(mockTr.readAll(any(byte[].class), any(int.class), any(int.class)))
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    buff[3] = 10;
                    return 4;
                }) // Read frame 1 size
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    for (int i = 0; i < 10; i++) {
                        buff[i] = 1;
                    }
                    return 10;
                }) // Read frame 1
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    buff[3] = 5;
                    return 4;
                }) // Read frame 2 size
                .then(invocationOnMock -> {
                    byte[] buff = (byte[]) invocationOnMock.getArguments()[0];
                    for (int i = 0; i < 5; i++) {
                        buff[i] = 2;
                    }
                    return 5;
                }) // Read frame 2
                .thenThrow(new TTransportException(TTransportException.END_OF_FILE));
        Runnable reader = tr.newTransportReader();

        reader.run();

        assertFalse(tr.isOpen());
        verify(mockRegistry).execute(new byte[]{1, 1, 1, 1, 1, 1, 1, 1, 1, 1});
        verify(mockRegistry).execute(new byte[]{2, 2, 2, 2, 2});
    }

    /**
     * Ensures TransportReader reads frames from the transport and closes it when
     * an error is encountered.
     */
    @Test
    public void testTransportReader_error() throws TException {
        FRegistry mockRegistry = mock(FRegistry.class);
        tr.registry = mockRegistry;
        FAdapterTransport.ExecutorFactory mockExecutorFactory = mock(FAdapterTransport.ExecutorFactory.class);
        ExecutorService mockExecutor = mock(ExecutorService.class);
        when(mockExecutorFactory.newExecutor()).thenReturn(mockExecutor);
        tr.setExecutorFactory(mockExecutorFactory);
        tr.open();
        when(mockTr.isOpen()).thenReturn(true);
        TTransportException cause = new TTransportException(TTransportException.UNKNOWN, "error");
        when(mockTr.readAll(any(byte[].class), any(int.class), any(int.class))).thenThrow(cause);
        Runnable reader = tr.newTransportReader();

        reader.run();

        assertFalse(tr.isOpen());
    }

    /**
     * Ensures request calls through to write and flush the underlying transport.
     */
    @Test
    public void testRequest() throws TTransportException {
        byte[] expectedResponse = "hi".getBytes();
        tr.registry = new MockRegistry(expectedResponse);
        when(mockTr.isOpen()).thenReturn(true);
        Mockito.doNothing().when(mockTr).open();
        tr.open();

        FContext context = new FContext();
        byte[] buff = "helloworld".getBytes();
        byte[] actualResponse = tr.request(context, false, buff);
        assertArrayEquals(expectedResponse, actualResponse);

        verify(mockTr).write(buff);
        verify(mockTr).flush();
    }

    /**
     * Ensures request throws TTransportException if the transport is not open.
     */
    @Test(expected = TTransportException.class)
    public void test_notOpen() throws TTransportException {
        tr.request(null, false, new byte[0]);
    }


    class MockRegistry implements FRegistry {

        byte[] response;

        MockRegistry(byte[] response) {
            this.response = response;
        }

        /**
         * @param context
         * @throws TTransportException if the given context is already registered to a callback.
         */
        @Override
        public void assignOpId(FContext context) throws TTransportException {

        }

        /**
         * Register a queue for the given FContext.
         *
         * @param context the FContext to register.
         * @param queue   the queue to place responses directed at this context.
         */
        @Override
        public void register(FContext context, BlockingQueue<byte[]> queue) {
            try {
                queue.put(response);
            } catch (Exception ignored) {
            }
        }

        /**
         * Unregister the callback for the given FContext.
         *
         * @param context the FContext to unregister.
         */
        @Override
        public void unregister(FContext context) {

        }

        /**
         * Dispatch a single Frugal message frame.
         *
         * @param frame an entire Frugal message frame.
         * @throws TException if execution failed.
         */
        @Override
        public void execute(byte[] frame) throws TException {

        }

        /**
         * Interrupt any registered contexts.
         */
        @Override
        public void close() {

        }
    }
}

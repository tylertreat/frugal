package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocol;
import com.workiva.frugal.protocol.FProtocolFactory;
import io.nats.client.Connection;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import io.nats.client.Subscription;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TJSONProtocol;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;

import java.io.IOException;
import java.util.Arrays;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.TimeUnit;

import static org.junit.Assert.*;
import static org.mockito.Mockito.*;

public class FStatelessNatsServerTest {

    private Connection conn;
    private FProcessor processor;
    private FProtocolFactory protocolFactory;
    private String subject = "foo";
    private String queue = "bar";
    private FStatelessNatsServer server;

    @Before
    public void setUp() {
        conn = mock(Connection.class);
        processor = mock(FProcessor.class);
        protocolFactory = mock(FProtocolFactory.class);
        server = new FStatelessNatsServer.Builder(conn, processor, protocolFactory, subject).withQueueGroup(queue).build();
    }

    @Test
    public void testServe() throws TException, IOException, InterruptedException {
        ArgumentCaptor<String> subjectCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<String> queueCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        Subscription sub = mock(Subscription.class);
        when(conn.subscribe(subjectCaptor.capture(), queueCaptor.capture(), handlerCaptor.capture())).thenReturn(sub);

        final BlockingQueue<Object> wait = new SynchronousQueue<>();

        new Thread(new Runnable() {
            public void run() {
                try {
                    server.serve();
                    wait.put(new Object());
                } catch (Exception e) {
                    fail(e.getMessage());
                }
            }
        }).start();

        server.stop();

        if (wait.poll(1, TimeUnit.SECONDS) == null) {
            fail("Wait timed out");
        }

        assertEquals(subject, subjectCaptor.getValue());
        assertEquals(queue, queueCaptor.getValue());
        assertNotNull(handlerCaptor.getValue());
        verify(sub).unsubscribe();
    }

    @Test
    public void testRequestHandler() {
        ExecutorService executor = mock(ExecutorService.class);
        ArgumentCaptor<Runnable> captor = ArgumentCaptor.forClass(Runnable.class);
        when(executor.submit(captor.capture())).thenReturn(null);
        server.workerPool = executor;
        MessageHandler handler = server.newRequestHandler();
        String reply = "reply";
        byte[] data = "this is a request".getBytes();
        Message msg = new Message(subject, reply, data);

        handler.onMessage(msg);

        verify(executor).submit(captor.getValue());
        assertEquals(FStatelessNatsServer.Request.class, captor.getValue().getClass());
        FStatelessNatsServer.Request request = (FStatelessNatsServer.Request) captor.getValue();
        assertArrayEquals(data, request.frameBytes);
        assertEquals(reply, request.reply);
        assertEquals(5000, request.highWatermark);
        assertEquals(protocolFactory, request.inputProtoFactory);
        assertEquals(protocolFactory, request.outputProtoFactory);
        assertEquals(processor, request.processor);
        assertEquals(conn, request.conn);
    }

    @Test
    public void testRequestHandler_noReply() {
        ExecutorService executor = mock(ExecutorService.class);
        when(executor.submit(any(Runnable.class))).thenReturn(null);
        server.workerPool = executor;
        MessageHandler handler = server.newRequestHandler();
        byte[] data = "this is a request".getBytes();
        Message msg = new Message(subject, null, data);

        handler.onMessage(msg);

        verify(executor, times(0)).submit(any(Runnable.class));
    }

    @Test
    public void testRequestProcess() throws TException, IOException {
        byte[] data = "xxxxhello".getBytes();
        long timestamp = System.currentTimeMillis();
        String reply = "reply";
        long highWatermark = 5000;
        MockFProcessor processor = new MockFProcessor(data, "blah".getBytes());
        protocolFactory = new FProtocolFactory(new TJSONProtocol.Factory());
        FStatelessNatsServer.Request request = new FStatelessNatsServer.Request(data, timestamp, reply, highWatermark,
                protocolFactory, protocolFactory, processor, conn);

        request.run();

        byte[] expected = new byte[]{0, 0, 0, 6, 34, 98, 108, 97, 104, 34};
        verify(conn).publish(reply, expected);
    }

    @Test
    public void testRequestProcess_noResponse() throws TException, IOException {
        byte[] data = "xxxxhello".getBytes();
        long timestamp = System.currentTimeMillis();
        String reply = "reply";
        long highWatermark = 5000;
        MockFProcessor processor = new MockFProcessor(data, null);
        protocolFactory = new FProtocolFactory(new TJSONProtocol.Factory());
        FStatelessNatsServer.Request request = new FStatelessNatsServer.Request(data, timestamp, reply, highWatermark,
                protocolFactory, protocolFactory, processor, conn);

        request.run();

        verify(conn, times(0)).publish(any(String.class), any(byte[].class));
    }

    private class MockFProcessor implements FProcessor {

        private byte[] expectedIn;
        private byte[] expectedOut;

        public MockFProcessor(byte[] expectedIn, byte[] expectedOut) {
            this.expectedIn = expectedIn;
            this.expectedOut = expectedOut;
        }

        @Override
        public void process(FProtocol in, FProtocol out) throws TException {
            assertTrue(in.getTransport() instanceof TMemoryInputTransport);

            if (expectedIn != null) {
                TMemoryInputTransport transport = (TMemoryInputTransport) in.getTransport();
                assertArrayEquals(Arrays.copyOfRange(expectedIn, 4, expectedIn.length), transport.getBuffer());
            }

            if (expectedOut != null) {
                out.writeString(new String(expectedOut));
            }
        }

    }

}

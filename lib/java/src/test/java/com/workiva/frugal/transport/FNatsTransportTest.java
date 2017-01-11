package com.workiva.frugal.transport;

import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FRegistry;
import io.nats.client.AsyncSubscription;
import io.nats.client.Connection;
import io.nats.client.Constants;
import io.nats.client.Message;
import io.nats.client.MessageHandler;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Mockito;

import java.io.IOException;
import java.util.concurrent.BlockingQueue;

import static com.workiva.frugal.transport.FNatsTransport.NATS_MAX_MESSAGE_SIZE;
import static org.junit.Assert.assertArrayEquals;
import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.fail;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Matchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FNatsTransport}.
 */
public class FNatsTransportTest {

    private Connection conn;
    private String subject = "foo";
    private String inbox = "bar";
    private FNatsTransport transport;

    @Before
    public void setUp() {
        conn = mock(Connection.class);
        transport = FNatsTransport.of(conn, subject).withInbox(inbox);
    }

    @Test(expected = TTransportException.class)
    public void testOpen_natsDisconnected() throws TTransportException {
        assertFalse(transport.isOpen());
        when(conn.getState()).thenReturn(Constants.ConnState.CLOSED);
        transport.open();
    }

    @Test
    public void testOpenCallbackClose() throws TException, IOException, InterruptedException {
        assertFalse(transport.isOpen());
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        ArgumentCaptor<String> inboxCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<MessageHandler> handlerCaptor = ArgumentCaptor.forClass(MessageHandler.class);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(inboxCaptor.capture(), handlerCaptor.capture())).thenReturn(sub);

        FRegistry mockRegistry = mock(FRegistry.class);
        transport.registry = mockRegistry;
        transport.open();

        verify(conn).subscribe(inboxCaptor.getValue(), handlerCaptor.getValue());
        assertEquals(inbox, inboxCaptor.getValue());

        MessageHandler handler = handlerCaptor.getValue();
        byte[] framedPayload = new byte[]{0, 1, 2, 3, 4, 5, 6, 7};
        byte[] payload = new byte[]{4, 5, 6, 7};
        Message msg = new Message("foo", "bar", framedPayload);
        handler.onMessage(msg);
        verify(mockRegistry).execute(payload);

        try {
            transport.open();
            fail("Expected TTransportException");
        } catch (TTransportException e) {
            assertEquals(TTransportException.ALREADY_OPEN, e.getType());
        }

        FTransportClosedCallback mockCallback = mock(FTransportClosedCallback.class);
        transport.setClosedCallback(mockCallback);
        transport.close();

        verify(sub).unsubscribe();
        verify(mockCallback).onClose(null);
        verify(mockRegistry).close();
    }

    @Test(expected = FMessageSizeException.class)
    public void testRequest_requestException() throws TTransportException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);
        transport.open();

        transport.request(new FContext(), false, new byte[NATS_MAX_MESSAGE_SIZE + 1]);
    }

    @Test
    public void testRequest() throws TTransportException, IOException, InterruptedException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        AsyncSubscription sub = mock(AsyncSubscription.class);
        when(conn.subscribe(any(String.class), any(MessageHandler.class))).thenReturn(sub);

        FRegistry mockRegistry = mock(FRegistry.class);
        transport.registry = mockRegistry;
        transport.open();

        FContext context = new FContext();
        byte[] buff = "helloworld".getBytes();
        byte[] expectedResponse = "hi".getBytes();
        Mockito.doAnswer(invocation -> {
            Object[] arg = invocation.getArguments();
            BlockingQueue<byte[]> queue = (BlockingQueue<byte[]>) arg[1];
            queue.put(expectedResponse);
            return null;
        }).when(mockRegistry).register(eq(context), any());
        byte[] actualResponse = transport.request(context, false, buff);

        assertArrayEquals(expectedResponse, actualResponse);
        verify(conn).publish(subject, inbox, buff);
    }

    @Test(expected = TTransportException.class)
    public void testRequest_notOpen() throws TTransportException {
        when(conn.getState()).thenReturn(Constants.ConnState.CONNECTED);
        transport.request(new FContext(), false, "helloworld".getBytes());
    }
}

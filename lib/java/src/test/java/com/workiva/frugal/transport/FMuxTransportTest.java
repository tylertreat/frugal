package com.workiva.frugal.transport;

import com.workiva.frugal.protocol.FClientRegistry;
import com.workiva.frugal.protocol.FRegistry;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.junit.Before;
import org.junit.Test;

import static org.junit.Assert.*;
import static org.mockito.Mockito.*;

public class FMuxTransportTest {

    private FMuxTransport muxTransport;
    private TTransport mockTrans;
    private FRegistry registry;

    @Before
    public void setUp() throws Exception {
        mockTrans = mock(TTransport.class);
        muxTransport = new FMuxTransport.Factory(4).getTransport(mockTrans);
    }

    @Test
    public void testIsOpenTrue() throws Exception {
        when(mockTrans.isOpen()).thenReturn(true);

        registry = new FClientRegistry();

        muxTransport.setRegistry(registry);

        assertTrue(muxTransport.isOpen());
    }

    @Test
    public void testIsOpenFalseWhenTransportClosed() throws Exception {
        when(mockTrans.isOpen()).thenReturn(false);

        registry = new FClientRegistry();

        muxTransport.setRegistry(registry);

        assertFalse(muxTransport.isOpen());
    }

    @Test
    public void testCloseCleanCloseNotOpen() {
        when(mockTrans.isOpen()).thenReturn(false);

        muxTransport.close();

        verify(mockTrans, times(0)).close();
    }

    @Test
    public void testCloseCleanClose() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(true);
        FRegistry mockRegistry = mock(FRegistry.class);
        muxTransport.setRegistry(mockRegistry);
        muxTransport.open();

        muxTransport.close();

        verify(mockTrans).close();
        verify(mockRegistry).close();
    }

    @Test
    public void testCloseUncleanCloseNotOpen() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(false);
        FRegistry mockRegistry = mock(FRegistry.class);
        muxTransport.setRegistry(mockRegistry);
        muxTransport.open();

        muxTransport.close(new Exception());

        verify(mockTrans).close();
        verify(mockRegistry).close();
    }

    @Test
    public void testCloseUncleanClose() throws TTransportException {
        when(mockTrans.isOpen()).thenReturn(true);
        FRegistry mockRegistry = mock(FRegistry.class);
        muxTransport.setRegistry(mockRegistry);
        muxTransport.open();

        muxTransport.close(new Exception());

        verify(mockTrans).close();
        verify(mockRegistry).close();
    }
}

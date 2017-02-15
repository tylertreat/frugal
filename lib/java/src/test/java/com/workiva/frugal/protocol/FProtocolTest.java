package com.workiva.frugal.protocol;

import com.workiva.frugal.FContext;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TProtocol;
import org.apache.thrift.transport.TMemoryBuffer;
import org.apache.thrift.transport.TTransport;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;


import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FProtocol}.
 */
@RunWith(JUnit4.class)
public class FProtocolTest {

    private TTransport transport;
    private TProtocol wrapped;
    private FProtocol protocol;
    private FContext context;

    @Before
    public void setUp() throws Exception {
        transport = mock(TTransport.class);
        wrapped = mock(TProtocol.class);
        protocol = new FProtocol(wrapped);
        context = new FContext("cId");
    }

    @Test
    public void testWriteRequestHeaders() throws Exception {
        when(wrapped.getTransport()).thenReturn(transport);

        protocol.writeRequestHeader(context);

        verify(transport).write(HeaderUtils.encode(context.getRequestHeaders()));
    }

    @Test
    public void testReadRequestHeaders() throws Exception {
        TMemoryBuffer memoryBuffer = new TMemoryBuffer(1024);
        FProtocol binaryProtocol = new FProtocol(new TBinaryProtocol(memoryBuffer));
        memoryBuffer.write(HeaderUtils.encode(context.getRequestHeaders()));

        FContext ctx = binaryProtocol.readRequestHeader();
        assertEquals(context.getCorrelationId(), ctx.getResponseHeader(FContext.CID_HEADER));
    }

}

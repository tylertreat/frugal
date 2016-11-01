package com.workiva.frugal.processor;

import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.FContext;
import com.workiva.frugal.protocol.FProtocol;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.protocol.TField;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.transport.TTransport;
import org.junit.Before;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.assertEquals;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

/**
 * Tests for {@link FBaseProcessor}.
 */
@RunWith(JUnit4.class)
public class FBaseProcessorTest {

    private final String oneWay = "oneWay";
    private HashMap<String, FProcessorFunction> map;
    private FProcessorFunction oneWayFunction;
    private FBaseProcessor processor;
    private FProtocol iprot;
    private FProtocol oprot;

    @Before
    public void setUp() throws Exception {
        map = new HashMap<>();
        oneWayFunction = mock(FProcessorFunction.class);

        processor = new TestFProcessor(map);

        iprot = mock(FProtocol.class);
        oprot = mock(FProtocol.class);
    }

    @Test
    public void testProcess() throws Exception {

        map.put(oneWay, oneWayFunction);

        FContext ctx = new FContext();
        TMessage thriftMessage = new TMessage(oneWay, (byte) 0x00, 1);

        when(iprot.readRequestHeader()).thenReturn(ctx);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);

        processor.process(iprot, oprot);

        verify(oneWayFunction).process(ctx, iprot, oprot);
    }

    @Test
    public void testProcessThrowsTApplicationException() throws Exception {
        TField tField = mock(TField.class);
        when(iprot.readFieldBegin()).thenReturn(tField);

        FContext ctx = new FContext();
        when(iprot.readRequestHeader()).thenReturn(ctx);

        TMessage thriftMessage = new TMessage("unknown", (byte) 0x00, 1);
        when(iprot.readMessageBegin()).thenReturn(thriftMessage);

        TTransport tTransport = mock(TTransport.class);
        when(oprot.getTransport()).thenReturn(tTransport);

        try {
            processor.process(iprot, oprot);
        } catch (TApplicationException ex) {
            assertEquals("Unknown function unknown", ex.getMessage());
            verify(oprot).writeResponseHeader(ctx);
            verify(oprot).writeMessageBegin(any(TMessage.class));
            verify(oprot).writeMessageEnd();
            verify(tTransport).flush();
        }
    }

    private class TestFProcessor extends FBaseProcessor {

        private Map<String, FProcessorFunction> procMap;

        public TestFProcessor(Map<String, FProcessorFunction> procMap) {
            this.procMap = procMap;
        }

        @Override
        public Map<String, FProcessorFunction> getProcessMap() {
            return procMap;
        }

        @Override
        public void addMiddleware(ServiceMiddleware middleware) {
        }
    }
}

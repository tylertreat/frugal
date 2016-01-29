package com.workiva.frugal.processor;

import com.workiva.frugal.FContext;
import com.workiva.frugal.FProtocol;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.protocol.TMessageType;
import org.apache.thrift.protocol.TProtocolUtil;
import org.apache.thrift.protocol.TType;

import java.util.Map;

public class FBaseProcessor implements FProcessor {

    protected static final Object WRITE_LOCK = new Object();

    private final Map<String, FProcessorFunction> processMap;

    protected FBaseProcessor(Map<String, FProcessorFunction> processorFunctionMap) {
        this.processMap = processorFunctionMap;
    }

    @Override
    public void process(FProtocol iprot, FProtocol oprot) throws TException {
        FContext ctx = iprot.readRequestHeader();
        TMessage message = iprot.readMessageBegin();
        FProcessorFunction processor = this.processMap.get(message.name);
        if (processor != null) {
            processor.process(ctx, iprot, oprot);
            return;
        }
        TProtocolUtil.skip(iprot, TType.STRUCT);
        iprot.readMessageEnd();
        TApplicationException e = new TApplicationException(TApplicationException.UNKNOWN_METHOD, "Unknown function " + message.name);
        synchronized (WRITE_LOCK) {
            oprot.writeResponseHeader(ctx);
            oprot.writeMessageBegin(new TMessage(message.name, TMessageType.EXCEPTION, 0));
            e.write(oprot);
            oprot.writeMessageEnd();
            oprot.getTransport().flush();
        }
        throw e;
    }

}

package com.workiva.frugal.processor;

import org.apache.thrift.transport.TTransport;

/**
 * FProcessorFactory creates FProcessors. The default factory just returns a singleton instance.
 */
public class FProcessorFactory {

    private FProcessor processor;

    public FProcessorFactory(FProcessor processor) {
        this.processor = processor;
    }

    public FProcessor getProcessor(TTransport trans) {
        return processor;
    }

}

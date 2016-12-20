package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;

import java.io.IOException;
import java.util.Arrays;

/**
 * Helper class to process Frugal frames.
 */
public class FrameProcessor {

    private final FProcessor processor;
    private final FProtocolFactory inProtocolFactory;
    private final FProtocolFactory outProtocolFactory;
    private final StringBuilder requestBuffer = new StringBuilder();

    private FrameProcessor(FProcessor processor,
                           FProtocolFactory inProtocolFactory,
                           FProtocolFactory outProtocolFactory) {
        this.processor = processor;
        this.inProtocolFactory = inProtocolFactory;
        this.outProtocolFactory = outProtocolFactory;
    }

    /**
     * Create a new FrameProcessor, setting the input and output protocol.
     *
     * @param processor Frugal request processor
     * @param protocolFactory input and output protocol
     * @return a new processor
     */
    public static FrameProcessor of(FProcessor processor,
                                    FProtocolFactory protocolFactory) {
        return new FrameProcessor(processor, protocolFactory, protocolFactory);
    }

    /**
     * Create a new FrameProcessor, setting the input and output protocol.
     *
     * @param processor Frugal request processor
     * @param inProtocolFactory input protocol
     * @param outProtocolFactory output protocol
     * @return a new processor
     */
    public static FrameProcessor of(FProcessor processor,
                                       FProtocolFactory inProtocolFactory,
                                       FProtocolFactory outProtocolFactory) {
        return new FrameProcessor(processor, inProtocolFactory, outProtocolFactory);
    }

    /**
     * Process one frame of data.
     *
     * @param data a raw input frame
     * @return The processed frame
     * @throws TException if error processing frame
     */
    protected String process(String data) throws TException, IOException {
        byte[] outputBytes = process(data.getBytes());
        return new String(outputBytes);
    }

    /**
     * Process one frame of data.
     *
     * @param data a raw input frame
     * @return The processed frame
     * @throws TException if error processing frame
     * @throws IOException if invalid request frame
     */
    protected byte[] process(byte[] data) throws TException, IOException {
        byte[] inputBytes = Base64.decodeBase64(data);
        if (inputBytes.length <= 4) {
            throw new IOException("Invalid request frame.");
        }

        // Exclude first 4 bytes (frame size)
        byte[] inputFrame = Arrays.copyOfRange(inputBytes, 4, inputBytes.length);

        // Run processor
        TTransport inTransport = new TMemoryInputTransport(inputFrame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();
        processor.process(inProtocolFactory.getProtocol(inTransport), outProtocolFactory.getProtocol(outTransport));

        return Base64.encodeBase64(outTransport.getWriteBytes());
    }
}

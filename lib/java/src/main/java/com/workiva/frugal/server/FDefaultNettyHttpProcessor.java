package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.DefaultHttpHeaders;
import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpHeaders;
import io.netty.handler.codec.http.HttpUtil;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Map;

import static io.netty.handler.codec.http.HttpHeaderNames.CONTENT_LENGTH;
import static io.netty.handler.codec.http.HttpHeaderNames.CONTENT_TRANSFER_ENCODING;
import static io.netty.handler.codec.http.HttpHeaderNames.CONTENT_TYPE;
import static io.netty.handler.codec.http.HttpHeaderNames.DATE;
import static io.netty.handler.codec.http.HttpResponseStatus.CONTINUE;
import static io.netty.handler.codec.http.HttpResponseStatus.INTERNAL_SERVER_ERROR;
import static io.netty.handler.codec.http.HttpResponseStatus.OK;
import static io.netty.handler.codec.http.HttpResponseStatus.REQUEST_ENTITY_TOO_LARGE;
import static io.netty.handler.codec.http.HttpVersion.HTTP_1_1;

/**
 * Default processor implementation for {@link FNettyHttpProcessor}.
 */
public class FDefaultNettyHttpProcessor implements FNettyHttpProcessor {

    private static final Logger LOGGER = LoggerFactory.getLogger(FNatsServer.class);

    private final FProcessor processor;
    private final FProtocolFactory inProtocolFactory;
    private final FProtocolFactory outProtocolFactory;
    private final Collection<Map.Entry<String, String>> customHeaders;

    private FDefaultNettyHttpProcessor(
            FProcessor processor,
            FProtocolFactory inProtocolFactory,
            FProtocolFactory outProtocolFactory) {
        this.processor = processor;
        this.inProtocolFactory = inProtocolFactory;
        this.outProtocolFactory = outProtocolFactory;
        this.customHeaders = new ArrayList<>();
    }

    /**
     * Create a new HTTP processer, setting the input and output protocol.
     *
     * @param processor       Frugal request processor
     * @param protocolFactory input and output protocol
     * @return a new processor
     */
    public static FDefaultNettyHttpProcessor of(FProcessor processor, FProtocolFactory protocolFactory) {
        return new FDefaultNettyHttpProcessor(processor, protocolFactory, protocolFactory);
    }

    /**
     * Create a new HTTP processor, setting the input and output protocol.
     *
     * @param processor          Frugal request processor
     * @param inProtocolFactory  input protocol
     * @param outProtocolFactory output protocol
     * @return a new processor
     */
    public static FDefaultNettyHttpProcessor of(
            FProcessor processor,
            FProtocolFactory inProtocolFactory,
            FProtocolFactory outProtocolFactory) {
        return new FDefaultNettyHttpProcessor(processor, inProtocolFactory, outProtocolFactory);
    }

    /**
     * Returns the size limit of the response payload.
     * Set in the x-frugal-payload-limit HTTP header.
     *
     * @param headers HTTP headers from incoming request
     * @return The size limit of the response, 0 if no limit header set
     */
    protected Integer getResponseLimit(HttpHeaders headers) {
        String payloadHeader = headers.get("x-frugal-payload-limit");
        Integer responseLimit;
        try {
            responseLimit = Integer.parseInt(payloadHeader);
        } catch (NumberFormatException ignored) {
            responseLimit = 0;
        }
        return responseLimit;
    }

    /**
     * Process one frame of data.
     *
     * @param inputBuffer an input frame
     * @return The processes frame as an output buffer
     * @throws TException if error processing frame
     */
    protected ByteBuf processFrame(ByteBuf inputBuffer) throws TException {
        byte[] encodedBytes = new byte[inputBuffer.readableBytes()];
        inputBuffer.readBytes(encodedBytes);
        byte[] inputBytes = Base64.decodeBase64(encodedBytes);

        // Exclude first 4 bytes which represent frame size
        byte[] inputFrame = Arrays.copyOfRange(inputBytes, 4, inputBytes.length);

        TTransport inTransport = new TMemoryInputTransport(inputFrame);
        TMemoryOutputBuffer outTransport = new TMemoryOutputBuffer();

        processor.process(inProtocolFactory.getProtocol(inTransport), outProtocolFactory.getProtocol(outTransport));

        byte[] outputBytes = Base64.encodeBase64(outTransport.getWriteBytes());
        return Unpooled.copiedBuffer(outputBytes);
    }

    /**
     * Process one frame of data.
     *
     * @param request an HTTP request
     * @return an HTTP response, processed by an FProcessor
     */
    @Override
    public FullHttpResponse process(FullHttpRequest request) {
        if (HttpUtil.is100ContinueExpected(request)) {
            return new DefaultFullHttpResponse(HTTP_1_1, CONTINUE);
        }

        ByteBuf body = request.content();
        ByteBuf outputBuffer = Unpooled.EMPTY_BUFFER;
        try {
            outputBuffer = processFrame(body);
        } catch (TException e) {
            LOGGER.error("Frugal processor returned unhandled error:" + e);
            return new DefaultFullHttpResponse(
                    HTTP_1_1,
                    INTERNAL_SERVER_ERROR);
        }

        Integer responseLimit = getResponseLimit(request.headers());
        if (responseLimit > 0 && outputBuffer.readableBytes() > responseLimit) {
            LOGGER.error("Response size too large for client");
            return new DefaultFullHttpResponse(
                    HTTP_1_1,
                    REQUEST_ENTITY_TOO_LARGE);
        }

        FullHttpResponse response = new DefaultFullHttpResponse(
                HTTP_1_1,
                OK,
                outputBuffer);

        ZonedDateTime dateTime = ZonedDateTime.now();
        DateTimeFormatter formatter = DateTimeFormatter.RFC_1123_DATE_TIME;

        DefaultHttpHeaders headers = (DefaultHttpHeaders) response.headers();
        headers.set(DATE, dateTime.format(formatter));
        headers.set(CONTENT_TYPE, "application/x-frugal");
        headers.set(CONTENT_TRANSFER_ENCODING, "base64");
        headers.set(CONTENT_LENGTH, Integer.toString(outputBuffer.readableBytes()));

        // Add custom headers
        for (Map.Entry<String, String> header : this.customHeaders) {
            headers.set(header.getKey(), header.getValue());
        }

        return response;
    }

    /**
     * Add a custom header to the returned response.
     *
     * @param key   Header name
     * @param value Header value
     */
    public void addCustomHeader(final String key, final String value) {
        this.customHeaders.add(new Map.Entry<String, String>() {
            public String getKey() {
                return key;
            }

            public String getValue() {
                return value;
            }

            public String setValue(String value) {
                return null;
            }
        });
    }

    /**
     * Add a map of custom header to the returned response.
     *
     * @param headers Map of header name, header value pairs.
     */
    public void setCustomHeaders(Collection<Map.Entry<String, String>> headers) {
        this.customHeaders.clear();
        this.customHeaders.addAll(headers);
    }
}

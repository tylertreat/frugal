package com.workiva.frugal.server;

import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.protocol.FProtocolFactory;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.DefaultFullHttpResponse;
import io.netty.handler.codec.http.FullHttpResponse;
import io.netty.handler.codec.http.HttpContent;
import io.netty.handler.codec.http.HttpHeaderNames;
import io.netty.handler.codec.http.HttpResponseStatus;
import io.netty.handler.codec.http.HttpUtil;
import io.netty.handler.codec.http.HttpHeaderValues;
import io.netty.handler.codec.http.HttpHeaders;
import io.netty.handler.codec.http.HttpObject;
import io.netty.handler.codec.http.HttpRequest;
import io.netty.handler.codec.http.LastHttpContent;
import io.netty.handler.codec.http.cookie.Cookie;
import io.netty.handler.codec.http.cookie.ServerCookieDecoder;
import io.netty.handler.codec.http.cookie.ServerCookieEncoder;
import io.netty.util.CharsetUtil;
import org.apache.commons.codec.binary.Base64;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;

import java.io.IOException;
import java.util.Arrays;
import java.util.Set;

import static io.netty.handler.codec.http.HttpResponseStatus.*;
import static io.netty.handler.codec.http.HttpVersion.*;

public class FNettyHttpHandler extends SimpleChannelInboundHandler<Object> {

    private HttpRequest request;
    private Integer responseLimit;

    private final FrameProcessor frameProcessor;
    private final StringBuilder requestBuffer = new StringBuilder();

    private FNettyHttpHandler(FProcessor processor,
                              FProtocolFactory inProtocolFactory,
                              FProtocolFactory outProtocolFactory) {
        this.frameProcessor = FrameProcessor.of(processor, inProtocolFactory, outProtocolFactory);
    }

    public static FNettyHttpHandler of(FProcessor processor, FProtocolFactory protocolFactory) {
        return new FNettyHttpHandler(processor, protocolFactory, protocolFactory);
    }

    public static FNettyHttpHandler of(FProcessor processor,
                                       FProtocolFactory inProtocolFactory,
                                       FProtocolFactory outProtocolFactory) {
        return new FNettyHttpHandler(processor, inProtocolFactory, outProtocolFactory);
    }

    @Override
    public void channelReadComplete(ChannelHandlerContext ctx) {
        ctx.flush();
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, Object msg) {
        if (msg instanceof HttpRequest) {
            HttpRequest request = this.request = (HttpRequest) msg;

            if (HttpUtil.is100ContinueExpected(request)) {
                send100Continue(ctx);
            }

            responseLimit = getResponseLimit(request.headers());
        }

        if (msg instanceof HttpContent) {
            HttpContent httpContent = (HttpContent) msg;

            ByteBuf content = httpContent.content();
            if (content.isReadable()) {
                requestBuffer.append(content.toString(CharsetUtil.UTF_8));
            }

            if (msg instanceof LastHttpContent) {

                LastHttpContent trailer = (LastHttpContent) msg;
                if (!writeResponse(trailer, ctx)) {
                    // If keep-alive is off, close the connection once the content is fully written.
                    ctx.writeAndFlush(Unpooled.EMPTY_BUFFER).addListener(ChannelFutureListener.CLOSE);
                }
            }
        }
    }

    /**
     * Process the request and write the result.
     *
     * @param currentObj
     * @param ctx
     * @return
     */
    protected boolean writeResponse(HttpObject currentObj, ChannelHandlerContext ctx) {

        // Process the frame, and set HTTP status
        HttpResponseStatus status = currentObj.decoderResult().isSuccess() ? OK : BAD_REQUEST;
        byte[] responseBytes = new byte[0];
        try {
            frameProcessor.process(requestBuffer.toString().getBytes());
            // Check size
            if (responseBytes.length > responseLimit) {
                status = REQUEST_ENTITY_TOO_LARGE;
            }

        } catch (TException | IOException e) {
            status = BAD_REQUEST;
        }

        // Build the response object.
        FullHttpResponse response = new DefaultFullHttpResponse(
                HTTP_1_1,
                status,
                Unpooled.copiedBuffer(responseBytes));

        // Set response headers
        response.headers().set(HttpHeaderNames.CONTENT_TYPE, "application/x-frugal");
        response.headers().set(HttpHeaderNames.CONTENT_TRANSFER_ENCODING, "base64");
        response.headers().setInt(HttpHeaderNames.CONTENT_LENGTH, response.content().readableBytes());

        // Decide whether to close the connection or not.
        boolean keepAlive = HttpUtil.isKeepAlive(request);
        if (keepAlive) {
            // Add keep alive header as per:
            response.headers().set(HttpHeaderNames.CONNECTION, HttpHeaderValues.KEEP_ALIVE);
        }

        // Write the response.
        ctx.write(response);

        return keepAlive;
    }

    private static void send100Continue(ChannelHandlerContext ctx) {
        FullHttpResponse response = new DefaultFullHttpResponse(HTTP_1_1, CONTINUE);
        ctx.write(response);
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


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        cause.printStackTrace();
        ctx.close();
    }
}
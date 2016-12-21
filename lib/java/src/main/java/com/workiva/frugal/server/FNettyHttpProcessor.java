package com.workiva.frugal.server;

import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;

/**
 * Processes a {@link FullHttpRequest} to return a {@link FullHttpResponse}.
 */
public interface FNettyHttpProcessor {

    /**
     * Process an HTTP request and return an HTTP response.
     *
     * @param request an HTTP request
     * @return a valid Frugal HTTP response
     */
    FullHttpResponse process(FullHttpRequest request);
}

package com.workiva.frugal.server;

import io.netty.handler.codec.http.FullHttpRequest;
import io.netty.handler.codec.http.FullHttpResponse;

/**
 * Processes a {@link FullHttpRequest} to return a {@link FullHttpResponse}.
 */
public interface FNettyHttpProcessor {

    /**
     * Process one frame of data.
     *
     * @param request a HTML request
     * @return The processed frame
     */
    FullHttpResponse process(FullHttpRequest request);
}

package com.workiva.frugal.protocol;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

/**
 * FContext is the context for a Frugal message. Every RPC has an FContext, which
 * can be used to set request headers, response headers, and the request timeout.
 * The default timeout is five seconds. An FContext is also sent with every publish
 * message which is then received by subscribers.
 * <p/>
 * In addition to headers, the FContext also contains a correlation ID which can
 * be used for distributed tracing purposes. A random correlation ID is generated
 * for each FContext if one is not provided.
 * <p/>
 * FContext also plays a key role in Frugal's multiplexing support. A unique,
 * per-request operation ID is set on every FContext before a request is made.
 * This operation ID is sent in the request and included in the response, which is
 * then used to correlate a response to a request. The operation ID is an internal
 * implementation detail and is not exposed to the user.
 * <p/>
 * This object is not thread-safe.
 */
public class FContext {

    protected static final String CID_HEADER = "_cid";         // Header containing correlation id
    protected static final String OPID_HEADER = "_opid";       // Header containing op id (uint64 as string)
    protected static final String TIMEOUT_HEADER = "_timeout"; // Header containing request timeout (milliseconds as string)

    protected static final long DEFAULT_TIMEOUT = 5 * 1000;

    private Map<String, String> requestHeaders = new ConcurrentHashMap<>();
    private Map<String, String> responseHeaders = new ConcurrentHashMap<>();

    private FContext(Map<String, String> requestHeaders, Map<String, String> responseHeaders) {
        this.requestHeaders = requestHeaders;
        this.responseHeaders = responseHeaders;
    }

    /**
     * Creates a new FContext with a randomly generated correlation id for tracing purposes.
     */
    public FContext() {
        this(generateCorrelationId());
    }

    /**
     * Creates a new FContext with the given correlation id for tracing purposes.
     *
     * @param correlationId unique tracing identifier
     */
    public FContext(String correlationId) {
        requestHeaders.put(CID_HEADER, correlationId);
        requestHeaders.put(OPID_HEADER, "0");
        requestHeaders.put(TIMEOUT_HEADER, Long.toString(DEFAULT_TIMEOUT));

    }

    /**
     * Creates a new FContext with the given request headers.
     *
     * @param headers request headers
     * @return FContext
     */
    protected static FContext withRequestHeaders(Map<String, String> headers) {
        headers.computeIfAbsent(CID_HEADER, k -> generateCorrelationId());
        headers.computeIfAbsent(TIMEOUT_HEADER, k -> Long.toString(DEFAULT_TIMEOUT));
        return new FContext(headers, new HashMap<>());
    }

    private static String generateCorrelationId() {
        return UUID.randomUUID().toString().replace("-", "");
    }

    /**
     * Returns the correlation id for the FContext. This is used for distributed-tracing purposes.
     *
     * @return correlation id
     */
    public String getCorrelationId() {
        return requestHeaders.get(CID_HEADER);
    }

    /**
     * Returns the operation id for the FContext. This is a unique long per operation. This is protected as operation
     * ids are an internal implementation detail.
     *
     * @return operation id
     */
    protected long getOpId() {
        String opIdStr = requestHeaders.get(OPID_HEADER);
        if (opIdStr == null) {
            return 0;
        }
        return Long.valueOf(opIdStr);
    }

    /**
     * Sets the operation id on the FContext. The operation id is used to map responses to requests. This is protected
     * as operation ids are an internal implementation detail.
     *
     * @param opId the operation id to set
     */
    protected void setOpId(long opId) {
        requestHeaders.put(OPID_HEADER, Long.toString(opId));
    }

    /**
     * Adds a request header to the FContext for the given name. A header is a key-value pair. If a header with the name
     * is already present on the FContext, it will be replaced. The _opid and _cid headers are reserved. Returns the
     * same FContext to allow for call chaining.
     *
     * @param name  header name
     * @param value header value
     * @return FContext
     */
    public FContext addRequestHeader(String name, String value) {
        if (OPID_HEADER.equals(name) || CID_HEADER.equals(name)) {
            return this;
        }
        requestHeaders.put(name, value);
        return this;
    }

    /**
     * Adds request headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced. The _opid
     * and _cid headers are reserved. Returns the same FContext to allow for call chaining.
     *
     * @param headers headers to add to request headers
     * @return FContext
     */
    public FContext addRequestHeaders(Map<String, String> headers) {
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            addRequestHeader(pair.getKey(), pair.getValue());
        }
        return this;
    }

    /**
     * Adds a response header to the FContext for the given name. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced.
     * The _opid header is reserved. Returns the same FContext to allow for call chaining.
     *
     * @param name  header name
     * @param value header value
     * @return FContext
     */
    public FContext addResponseHeader(String name, String value) {
        if (OPID_HEADER.equals(name)) {
            return this;
        }
        responseHeaders.put(name, value);
        return this;
    }

    /**
     * Adds response headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced. The _opid
     * header is reserved. Returns the same FContext to allow for call chaining.
     *
     * @param headers headers to add to request headers
     * @return FContext
     */
    public FContext addResponseHeaders(Map<String, String> headers) {
        for (Map.Entry<String, String> pair : headers.entrySet()) {
            addResponseHeader(pair.getKey(), pair.getValue());
        }
        return this;
    }

    /**
     * Adds response headers to the FContext for the given headers map. A header is a key-value pair.
     * If a header with the name is already present on the FContext, it will be replaced.
     *
     * @param headers headers to add to request headers
     */
    protected void forceAddResponseHeaders(Map<String, String> headers) {
        responseHeaders.putAll(headers);
    }

    /**
     * Returns the request header with the given name. If no such header exists, null is returned.
     *
     * @param name header name
     * @return header value or null if it doesn't exist
     */
    public String getRequestHeader(String name) {
        return requestHeaders.get(name);
    }

    /**
     * Returns the response header with the given name. If no such header exists, null is returned.
     *
     * @param name header name
     * @return header value or null if it doesn't exist
     */
    public String getResponseHeader(String name) {
        return responseHeaders.get(name);
    }

    /**
     * Returns the request headers on the FContext.
     *
     * @return request headers map
     */
    public Map<String, String> getRequestHeaders() {
        return new HashMap<>(requestHeaders);
    }

    /**
     * Returns the response headers on the FContext.
     *
     * @return response headers map
     */
    public Map<String, String> getResponseHeaders() {
        return new HashMap<>(responseHeaders);
    }

    /**
     * Get the request timeout.
     *
     * @return the request timeout in milliseconds.
     */
    public long getTimeout() {
        return Long.parseLong(requestHeaders.getOrDefault(TIMEOUT_HEADER, "0"));
    }

    /**
     * Set the request timeout. Default is 5 seconds.
     *
     * @param timeout timeout for the request in milliseconds.
     */
    public void setTimeout(long timeout) {
        requestHeaders.put(TIMEOUT_HEADER, Long.toString(timeout));
    }

    protected void setResponseOpId(String opId) {
        responseHeaders.put(OPID_HEADER, opId);
    }
}

package com.workiva.frugal.protocol;

import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

import java.util.HashMap;
import java.util.Map;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertNotEquals;
import static org.junit.Assert.assertNull;

/**
 * Tests for {@link FContext}.
 */
@RunWith(JUnit4.class)
public class FContextTest {

    @Test
    public void testGenerateCorrelationId() {
        FContext ctx = new FContext();
        assertNotEquals("", ctx.getCorrelationId());
    }

    @Test
    public void testCorrelationId() {
        String correlationId = "abc";
        FContext ctx = new FContext(correlationId);
        assertEquals(correlationId, ctx.getCorrelationId());
    }

    @Test
    public void testRequestHeader() {
        FContext ctx = new FContext();
        assertEquals(ctx, ctx.addRequestHeader("foo", "bar"));
        assertEquals(ctx, ctx.addRequestHeader("_cid", "123"));
        assertEquals("bar", ctx.getRequestHeader("foo"));
        assertNull(ctx.getRequestHeader("blah"));
    }

    @Test
    public void testAddRequestHeaders() {
        FContext ctx = new FContext();
        Map<String, String> headers = new HashMap<>();
        headers.put("foo", "bar");
        headers.put("baz", "qux");
        headers.put("_cid", "123");
        assertEquals(ctx, ctx.addRequestHeaders(headers));
        assertEquals("bar", ctx.getRequestHeader("foo"));
        assertEquals("qux", ctx.getRequestHeader("baz"));
    }

    @Test
    public void testResponseHeader() {
        FContext ctx = new FContext();
        assertEquals(ctx, ctx.addResponseHeader("foo", "bar"));
        assertEquals(ctx, ctx.addResponseHeader("_opid", "1"));
        assertEquals("bar", ctx.getResponseHeader("foo"));
        assertNull(ctx.getResponseHeader("blah"));
    }

    @Test
    public void testAddResponseHeaders() {
        FContext ctx = new FContext();
        Map<String, String> headers = new HashMap<>();
        headers.put("foo", "bar");
        headers.put("baz", "qux");
        headers.put("_opid", "1");
        assertEquals(ctx, ctx.addResponseHeaders(headers));
        assertEquals("bar", ctx.getResponseHeader("foo"));
        assertEquals("qux", ctx.getResponseHeader("baz"));
    }

}

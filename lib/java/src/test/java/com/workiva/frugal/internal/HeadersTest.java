package com.workiva.frugal.internal;

import static org.junit.Assert.*;

import com.workiva.frugal.FException;
import org.apache.thrift.transport.TMemoryInputTransport;
import org.apache.thrift.transport.TTransport;
import org.junit.Rule;
import org.junit.Test;
import org.junit.rules.ExpectedException;

import java.util.HashMap;
import java.util.Map;

public class HeadersTest {

    private static final Map<String, String> HEADERS;
    static {
        HEADERS = new HashMap<>();
        HEADERS.put("foo", "bar");
        HEADERS.put("blah", "baz");
    }

    private static final byte[] LIST = new byte[] { 0, 0, 0, 0, 29, 0, 0, 0, 3, 102, 111, 111, 0, 0, 0, 3, 98, 97,
            114, 0, 0, 0, 4, 98, 108, 97, 104, 0, 0, 0, 3, 98, 97, 122 };

    @Rule
    public final ExpectedException thrown = ExpectedException.none();

    @Test
    public void testReadOutOfTransport() throws FException {
        TTransport transport = new TMemoryInputTransport(LIST);

        Map<String, String> decodedHeaders = Headers.read(transport);
        assertEquals(HEADERS, decodedHeaders);
    }

    @Test
    public void testReadThrowsFExceptionForUnsupportedVersion() throws FException {
        TTransport transport = new TMemoryInputTransport(new byte[] {1});

        thrown.expect(FException.class);
        thrown.expectMessage("unsupported header version 1");
        Headers.read(transport);
    }

    @Test
    public void testReadThrowsFExceptionForTTransportException() throws FException {
        TTransport transport = new TMemoryInputTransport(new byte[] {0, 0, 0});

        thrown.expect(FException.class);
        thrown.expectMessage("could not read header version");
        Headers.read(transport);
    }

    @Test
    public void testDecodeFromFrame() throws FException {
        Map<String, String> decodedHeaders = Headers.decodeFromFrame(LIST);
        assertEquals(HEADERS, decodedHeaders);
    }

    @Test
    public void testEncodeDecode() throws FException {
        byte[] encodedHeaders = Headers.encode(HEADERS);
        Map<String, String> decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
        assertEquals(HEADERS, decodedHeaders);
    }

    @Test
    public void testEncodeDecodeNull() throws FException {
        Map<String, String> empty = new HashMap<>();
        byte[] encodedHeaders = Headers.encode(null);
        Map<String, String> decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
        assertEquals(empty, decodedHeaders);
    }

    @Test
    public void testEncodeDecodeEmpty() throws FException {
        Map<String, String> empty = new HashMap<>();
        byte[] encodedHeaders = Headers.encode(empty);
        Map<String, String> decodedHeaders = Headers.decodeFromFrame(encodedHeaders);
        assertEquals(empty, decodedHeaders);
    }

    @Test
    public void testDecodeHeadersFromFrameThrowsFExceptionForBadFrame() throws FException {
        thrown.expect(FException.class);
        thrown.expectMessage("invalid frame size 3");
        Headers.decodeFromFrame(new byte[3]);
    }

    @Test
    public void testDecodeHeadersFromFrameThrowsFExceptionForUnsupportedVersion() throws FException {
        thrown.expect(FException.class);
        thrown.expectMessage("unsupported header version 1");
        Headers.decodeFromFrame(new byte[] {1, 0, 0, 0, 0});
    }

}
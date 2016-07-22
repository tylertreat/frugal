package com.workiva;

import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TCompactProtocol;
import org.apache.thrift.protocol.TJSONProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.THttpClient;
import org.apache.thrift.transport.TSocket;
import org.apache.thrift.transport.TTransport;

import java.util.ArrayList;
import java.util.List;


public class utils {

    public static TProtocolFactory whichProtocolFactory (String protocol_type) throws Exception {
        List<String> validProtocols = new ArrayList<>();
        validProtocols.add("binary");
        validProtocols.add("compact");
        validProtocols.add("json");

        if (!validProtocols.contains(protocol_type)) {
            throw new Exception("Unknown protocol type! " + protocol_type);
        }

        TProtocolFactory protocolFactory;
        switch (protocol_type) {
            case "json":
                protocolFactory = new TJSONProtocol.Factory();
                break;
            case "compact":
                protocolFactory = new TCompactProtocol.Factory();
                break;
            case "binary":
                protocolFactory = new TBinaryProtocol.Factory();
                break;
            default:
                throw new Exception("Unknown protocol type encountered: " + protocol_type);
        }
        return protocolFactory;
    }

    public static TTransport whichTTransport (String transport_type, int socketTimeoutMs, String host, Integer port) throws Exception {
        List<String> validTransports = new ArrayList<>();
        validTransports.add("buffered");
        validTransports.add("framed");
        validTransports.add("http");

        if (!validTransports.contains(transport_type)) {
            throw new Exception("Unknown transport type! " + transport_type);
        }

        TTransport transport = null;
        try {
            if (transport_type.equals("http")) {
                String url = "http://" + host + ":" + port + "/service";
                transport = new THttpClient(url);
            } else {
                TSocket socket;
                socket = new TSocket(host, port);
                socket.setTimeout(socketTimeoutMs);
                transport = socket;
                switch (transport_type) {
                    case "buffered":
                        break;
                    case "framed":
                        break;
                }
            }
        } catch (Exception x) {
            x.printStackTrace();
            System.exit(1);
        }

        return transport;
    }
}

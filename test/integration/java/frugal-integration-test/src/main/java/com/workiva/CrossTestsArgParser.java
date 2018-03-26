package com.workiva;


public class CrossTestsArgParser {
    private String host = "localhost";
    private int port = 9090;
    private String protocolType = "binary";
    private String transportType = "nats";

    public CrossTestsArgParser(String[] cliArgs) {
        // TODO should use an actual arg parser
        try {
            for (String arg : cliArgs) {
                if (arg.startsWith("--host")) {
                    host = arg.split("=")[1];
                } else if (arg.startsWith("--port")) {
                    port = Integer.valueOf(arg.split("=")[1]);
                } else if (arg.startsWith("--protocol")) {
                    protocolType = arg.split("=")[1];
                } else if (arg.startsWith("--transport")) {
                    transportType = arg.split("=")[1];
                } else if (arg.equals("--help")) {
                    System.out.println("Allowed options:");
                    System.out.println("  --help\t\t\tProduce help message");
                    System.out.println("  --host=arg (=" + host + ")\tHost to connect, only for clients");
                    System.out.println("  --port=arg (=" + port + ")\tPort number to connect");
                    System.out.println("  --transport=arg (=" + transportType + ")\n\t\t\t\tTransport: nats, http");
                    System.out.println("  --protocol=arg (=" + protocolType + ")\tProtocol: binary, json, compact");
                    System.exit(0);
                }
            }
        } catch (Exception x) {
            System.err.println("Can not parse arguments! See --help");
            System.err.println("Exception parsing arguments: " + x);
            System.exit(1);
        }
    }

    public String getHost() {
        return host;
    }

    public int getPort() {
        return port;
    }

    public String getProtocolType() {
        return protocolType;
    }

    public String getTransportType() {
        return transportType;
    }
}

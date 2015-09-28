import com.workiva.frugal.NatsTransportFactory;
import com.workiva.frugal.TransportFactory;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransportFactory;
import org.nats.Connection;

import java.io.IOException;
import java.util.Properties;

public class Publisher {

    public static EventsPublisher publisher;

    public static void main(String[] args) throws IOException, InterruptedException, TException {
        Connection conn = Connection.connect(new Properties());
        TransportFactory tf = new NatsTransportFactory(conn);
        TTransportFactory thriftTf = new TTransportFactory();
        TProtocolFactory pf = new TBinaryProtocol.Factory();
        publisher = new EventsPublisher(tf, thriftTf, pf);
        for (int i = 0; i < 5; i++) {
            publisher.publishEventCreated(new Event(i, "Hello, world!"));
        }
        System.out.println("Published event");
    }

}

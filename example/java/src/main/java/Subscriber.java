import com.workiva.frugal.NatsTransportFactory;
import com.workiva.frugal.TransportFactory;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TProtocolFactory;
import org.apache.thrift.transport.TTransportFactory;
import org.nats.Connection;

import java.io.IOException;
import java.util.Properties;

public class Subscriber {

    public static EventsSubscriber subscriber;

    public static void main(String[] args) throws IOException, InterruptedException, TException {
        Connection conn = Connection.connect(new Properties());
        TransportFactory tf = new NatsTransportFactory(conn);
        TTransportFactory thriftTf = new TTransportFactory();
        TProtocolFactory pf = new TBinaryProtocol.Factory();
        subscriber = new EventsSubscriber(tf, thriftTf, pf);
        subscriber.subscribeEventCreated(new EventsSubscriber.EventCreatedHandler() {
            @Override
            public void onEventCreated(Event event) {
                System.out.println("received event: " + event.getID() + " " + event.getMessage());
            }
        });

        System.out.println("Subscriber started...");
        synchronized (subscriber) {
            subscriber.wait();
        }
    }

}

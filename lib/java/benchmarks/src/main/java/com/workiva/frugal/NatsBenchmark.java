package com.workiva.frugal;

import io.nats.client.Connection;
import io.nats.client.ConnectionFactory;
import org.openjdk.jmh.annotations.Benchmark;
import org.openjdk.jmh.annotations.BenchmarkMode;
import org.openjdk.jmh.annotations.Fork;
import org.openjdk.jmh.annotations.Measurement;
import org.openjdk.jmh.annotations.Mode;
import org.openjdk.jmh.annotations.Scope;
import org.openjdk.jmh.annotations.Setup;
import org.openjdk.jmh.annotations.State;
import org.openjdk.jmh.annotations.TearDown;
import org.openjdk.jmh.annotations.Threads;
import org.openjdk.jmh.annotations.Warmup;

import java.io.IOException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

@Warmup(iterations = 5, time = 1, timeUnit = TimeUnit.SECONDS)
@Measurement(iterations = 5, time = 1, timeUnit = TimeUnit.SECONDS)
@BenchmarkMode(Mode.Throughput)
@Threads(1)
@Fork(1)
@State(Scope.Thread)
public class NatsBenchmark {

    private static final int NUM_SUBSCRIBERS = 4;
    Connection nc;

    @Setup
    public void setup() throws IOException, TimeoutException {
        ConnectionFactory cf = new ConnectionFactory(ConnectionFactory.DEFAULT_URL);
        try {
            nc = cf.createConnection();
            for (int i = 0; i < NUM_SUBSCRIBERS; i++) {
                nc.subscribe("topic", m -> {
                    // do nothing
                });
            }
        } catch (IOException | TimeoutException e) {
            e.printStackTrace();
        }
    }

    @TearDown
    public void teardown() {
        nc.close();
    }

    @Benchmark
    public void testPublishToSub() {
        try {
            nc.publish("topic", "Hello World".getBytes());
        } catch (IOException e) {
            e.printStackTrace();
        }
    }
}

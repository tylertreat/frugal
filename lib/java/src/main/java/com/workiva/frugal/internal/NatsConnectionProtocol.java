package com.workiva.frugal.internal;

/**
 * Created by tylerrinnan on 1/29/16.
 */
public class NatsConnectionProtocol {
    public static final int NATS_V0 = 0;

    private int version;

    public NatsConnectionProtocol(int version){
        this.version = version;
    }

    public int getVersion() {
        return version;
    }

    public void setVersion(int version) {
        this.version = version;
    }

    @Override
    public String toString(){
        return String.format("Version: %d", this.version);
    }
}

package com.workiva;


import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelPipeline;
import io.netty.channel.socket.SocketChannel;
import io.netty.handler.codec.http.HttpRequestDecoder;
import io.netty.handler.codec.http.HttpResponseEncoder;

public class NettyHttpInitializer extends ChannelInitializer<SocketChannel> {

    FNettyHttpHandler handler;

    public NettyHttpInitializer(FNettyHttpHandler handler) {
        this.handler = handler;
    }

    @Override
    public void initChannel(SocketChannel ch) {
        ChannelPipeline p = ch.pipeline();
        p.addLast(new HttpRequestDecoder());
        p.addLast(new HttpResponseEncoder());
        p.addLast(handler);
    }
}
import 'dart:async';
import 'dart:convert' show BASE64;
import 'dart:convert' show Encoding;
import 'dart:convert' show Utf8Codec;
import 'dart:typed_data' show Uint8List;

import 'package:frugal/frugal.dart';
import 'package:http/http.dart' show BaseRequest;
import 'package:http/http.dart' show Client;
import 'package:http/http.dart' show Response;
import 'package:http/http.dart' show StreamedResponse;
import 'package:test/test.dart';
import 'package:thrift/thrift.dart';

void main() {
  const utf8Codec = const Utf8Codec();

  group('FHttpClientTransport', () {
    FakeHttpClient client;
    FakeFRegistry registry;
    FHttpClientTransport transport;

    Uint8List transportRequest = new Uint8List.fromList([1, 2, 3, 4, 5]);
    Uint8List transportRequestFramed =
        new Uint8List.fromList([0, 0, 0, 5, 1, 2, 3, 4, 5]);
    Uint8List transportResponse = new Uint8List.fromList([6, 7, 8, 9]);
    Uint8List transportResponseFramed =
        new Uint8List.fromList([0, 0, 0, 4, 6, 7, 8, 9]);

    setUp(() {
      client = new FakeHttpClient();
      var config = new FHttpConfig(Uri.parse('http://localhost'), {});
      transport = new FHttpClientTransport(client, config);
      registry = new FakeFRegistry();
      transport.setRegistry(registry);
    });

    test('Test transport sends body and receives response', () async {
      transport.writeAll(transportRequest);
      expect(client.postRequest, isEmpty);

      client.postResponse = BASE64.encode(transportResponseFramed);
      await transport.flush();

      expect(client.postRequest, isNotEmpty);

      var actualRequest = BASE64.decode(client.postRequest);
      expect(actualRequest, transportRequestFramed);

      expect(registry.data, transportResponse);
    });

    test('Test transport receives bad data', () async {
      client.postResponse = '`';
      transport.writeAll(transportRequest);
      expect(transport.flush(), throwsA(new isInstanceOf<TProtocolError>()));
    });
  });

  group('FHttpClientTransport request size too large', () {
    FakeHttpClient client;
    FakeFRegistry registry;
    FHttpClientTransport transport;

    setUp(() {
      client = new FakeHttpClient();
      var config = new FHttpConfig(Uri.parse('http://localhost'), {},
          requestSizeLimit: 10);
      transport = new FHttpClientTransport(client, config);
      registry = new FakeFRegistry();
      transport.setRegistry(registry);
    });

    test('Test transport receives error', () {
      expect(
          () => transport.writeAll(utf8Codec.encode('my really long request')),
          throwsA(new isInstanceOf<FMessageSizeError>()));
    });
  });

  group('FHttpClientTransport http post failed', () {
    FakeHttpClient client;
    FakeFRegistry registry;
    FHttpClientTransport transport;

    setUp(() {
      client = new FakeHttpClient(err: new StateError('baa!'));
      var config = new FHttpConfig(Uri.parse('http://localhost'), {});
      transport = new FHttpClientTransport(client, config);
      registry = new FakeFRegistry();
      transport.setRegistry(registry);
    });

    test('Test transport receives error', () async {
      var expectedText = 'my response';
      var expectedBytes = utf8Codec.encode(expectedText);
      client.postResponse = BASE64.encode(expectedBytes);

      transport.writeAll(utf8Codec.encode('my request'));
      expect(transport.flush(), throwsA(new isInstanceOf<TTransportError>()));
    });
  });

  group('FHttpClientTransport http response too large', () {
    FakeHttpClient client;
    FakeFRegistry registry;
    FHttpClientTransport transport;

    setUp(() {
      var expectedHeaders = {
        'X-Frugal-Payload-Limit': '10',
        'Content-Type': 'application/x-frugal',
        'Accept': 'application/x-frugal'
      };
      client = new FakeHttpClient(
          code: 413, sync: false, expectedHeaders: expectedHeaders);
      var config = new FHttpConfig(Uri.parse('http://localhost'), {},
          responseSizeLimit: 10);
      transport = new FHttpClientTransport(client, config);
      registry = new FakeFRegistry();
      transport.setRegistry(registry);
    });

    test('Test transport receives error', () async {
      var expectedText = 'my response';
      var expectedBytes = utf8Codec.encode(expectedText);
      client.postResponse = BASE64.encode(expectedBytes);

      transport.writeAll(utf8Codec.encode('my request'));
      expect(transport.flush(), throwsA(new isInstanceOf<FMessageSizeError>()));
    });
  });

  group('FHttpClientTransport http error code', () {
    FakeHttpClient client;
    FakeFRegistry registry;
    FHttpClientTransport transport;

    setUp(() {
      client = new FakeHttpClient(code: 404, sync: false);
      var config = new FHttpConfig(Uri.parse('http://localhost'), {});
      transport = new FHttpClientTransport(client, config);
      registry = new FakeFRegistry();
      transport.setRegistry(registry);
    });

    test('Test transport receives error', () async {
      var expectedText = 'my response';
      var expectedBytes = utf8Codec.encode(expectedText);
      client.postResponse = BASE64.encode(expectedBytes);

      transport.writeAll(utf8Codec.encode('my request'));
      expect(transport.flush(), throwsA(new isInstanceOf<TTransportError>()));
    });
  });
}

class FakeFRegistry extends FRegistry {
  Uint8List data;

  void register(FContext ctx, FAsyncCallback callback) {
    return;
  }

  void unregister(FContext ctx) {
    return;
  }

  void execute(Uint8List data) {
    this.data = data;
  }
}

class FakeHttpClient implements Client {
  String postResponse = '';
  String postRequest = '';
  Map expectedHeaders;

  final bool sync;
  final int code;
  final Error err;

  FakeHttpClient(
      {this.code: 200,
      this.sync: false,
      this.err: null,
      this.expectedHeaders: null});

  Future<Response> post(url,
      {Map<String, String> headers, body, Encoding encoding}) {
    if (err != null) throw err;

    if (expectedHeaders != null) {
      expectedHeaders.forEach((K, V) {
        if (headers[K] != V) {
          throw new Error();
        }
      });
    }

    postRequest = body;
    var response = new Response(postResponse, code);

    if (sync) {
      return new Future.sync(() => response);
    } else {
      return new Future.value(response);
    }
  }

  Future<Response> head(url, {Map<String, String> headers}) =>
      throw new UnimplementedError();

  Future<Response> get(url, {Map<String, String> headers}) =>
      throw new UnimplementedError();

  Future<Response> put(url,
          {Map<String, String> headers, body, Encoding encoding}) =>
      throw new UnimplementedError();

  Future<Response> patch(url,
          {Map<String, String> headers, body, Encoding encoding}) =>
      throw new UnimplementedError();

  Future<Response> delete(url, {Map<String, String> headers}) =>
      throw new UnimplementedError();

  Future<String> read(url, {Map<String, String> headers}) =>
      throw new UnimplementedError();

  Future<Uint8List> readBytes(url, {Map<String, String> headers}) =>
      throw new UnimplementedError();

  Future<StreamedResponse> send(BaseRequest request) =>
      throw new UnimplementedError();

  void close() => throw new UnimplementedError();
}

import "dart:typed_data";
import "package:test/test.dart";
import "package:frugal/frugal.dart";
import "package:thrift/thrift.dart";

void main() {
 test("writeRequestHeader writes the request headers and readRequestHeader reads the headers", () {
   var list = new Uint8List(1024);
   var transport = new TUint8List(list);
   var tProtocol = new TBinaryProtocol(transport);
   var fProtocol = new FProtocol(tProtocol);

   var ctx = new Context(correlationId: "sweet-corid");
   ctx.addRequestHeader("foo", "bar");
   fProtocol.writeRequestHeader(ctx);

   var decodedCtx = fProtocol.readRequestHeader();
   expect(decodedCtx.requestHeaders(), ctx.requestHeaders());
 });
}
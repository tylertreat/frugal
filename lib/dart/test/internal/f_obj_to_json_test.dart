import "package:test/test.dart";
import "package:frugal/frugal.dart";

void main() {
  group('fObjToJson', () {
    test('Serializes an FContext', () {
      String json = fObjToJson(new FContext(correlationId: "cid"));
      expect(json, '{"_cid":"cid","_opid":"0"}');
    });

    test('Serializes a normal object', () {
      String json = fObjToJson("foo");
      expect(json, '"foo"');
    });
  });
}

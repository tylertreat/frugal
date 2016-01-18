import "dart:typed_data";
import "package:test/test.dart";
import "package:frugal/frugal.dart";

var list = [0, 0, 0, 0, 29, 0, 0, 0, 3, 102, 111, 111, 0, 0, 0, 3, 98, 97,
114, 0, 0, 0, 4, 98, 108, 97, 104, 0, 0, 0, 3, 98, 97, 122];

void main() {
  test('test that write properly writes the bytes from the given buffer', () {
    var buffList = new Uint8List.fromList(list);
    var buff = new WriteByteBuffer(10);
    expect(10, buff.remaining);
    var n = buff.write(buffList, 0, buffList.length);
    expect(n, 10);
    var expected = new Uint8List.fromList(list.sublist(0, 10));
    expect(buff.asUint8List(), expected);
    expect(0, buff.remaining);
  });

  test('test that read properly reads the bytes into the given buffer', () {
    var buffList = new Uint8List.fromList(list);
    var buff = new ReadByteBuffer();
    buff.add(buffList);
    expect(list.length, buff.remaining);

    var readBuff = new Uint8List(10);
    var n = buff.read(readBuff, 0, 15);
    expect(10, n);
    var expected = new Uint8List.fromList(list.sublist(0, 10));
    expect(readBuff, expected);
    expect(list.length - 10, buff.remaining);
  });
}


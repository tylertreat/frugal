part of frugal.src.frugal;

final TSerializer _serializer =
    new TSerializer(protocolFactory: new TJsonProtocolFactory());

String fObjToJson(Object obj) {
  if (obj is TBase) {
    return new String.fromCharCodes(_serializer.write(obj));
  }
  if (obj is FContext) {
    return JSON.encode(obj.requestHeaders());
  }
  return JSON.encode(obj);
}

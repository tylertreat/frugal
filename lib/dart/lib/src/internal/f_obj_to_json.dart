part of frugal;

String fObjToJson(Object obj) {
  if (obj is TBase) {
    TSerializer serializer =
        new TSerializer(protocolFactory: new TJsonProtocolFactory());
    return new String.fromCharCodes(serializer.write(obj));
  }
  if (obj is FContext) {
    return JSON.encode(obj.requestHeaders());
  }
  return JSON.encode(obj);
}

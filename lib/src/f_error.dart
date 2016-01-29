part of frugal;

class FError extends TError {

  FError()
    : super(0, "");

  FError.withMessage(String message)
    : super(0, message);

  FError.withType(int type, String message)
    : super(type, message);
}
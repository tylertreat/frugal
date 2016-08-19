import webapp2

from frugal.server.http_handler import _FHttpRequest
from frugal.server.http_handler import _FSynchronousHttpRequestHandler


def new_webapp2_handler(processor, protocol_factory):
    """
    Produces a class extending webapp2.RequestHandler that can be used to handle
    frugal HTTP rpc requests.

    Args:
        processor: The processor to use to handle requests.
        protocol_factory: A protocol factory to serialize/deserialize frugal
                          requests.
    """
    handler = _FSynchronousHttpRequestHandler(processor, protocol_factory)

    class FWebapp2Handler(webapp2.RequestHandler):
        """
        FWebapp2Handler uses the webapp2 framework to handle frugal HTTP rpc
        requests.
        """
        def post(self):
            request = _FHttpRequest(
                headers=self.request.headers,
                body=self.request.body,
            )
            response = handler.handle_http_request(request)
            self.response.set_status(response.status_code)
            self.response.headers.update(response.headers)
            self.response.write(response.body)

    return FWebapp2Handler

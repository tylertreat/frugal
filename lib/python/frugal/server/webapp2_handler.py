import webapp2

from frugal.server.http_handler import FrugalHttpRequest
from frugal.server.http_handler import FrugalHttpRequestHandler


def new_webapp2_handler(processor, protocol_factory):
    handler = FrugalHttpRequestHandler(processor, protocol_factory)

    class FWebapp2Handler(webapp2.RequestHandler):
        def post(self):
            request = FrugalHttpRequest(
                headers=self.request.headers,
                body=self.request.body,
            )
            response = handler.handle_http_request(request)
            self.response.set_status(response.status_code)
            self.response.headers.update(response.headers)
            self.response.write(response.body)

    return FWebapp2Handler

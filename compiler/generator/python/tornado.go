package python

import (
	"fmt"
	"os"

	"github.com/Workiva/frugal/compiler/globals"
	"github.com/Workiva/frugal/compiler/parser"
)

// TornadoGenerator implements the LanguageGenerator interface for Python using
// Tornado.
type TornadoGenerator struct {
	*Generator
}

// GenerateServiceImports generates necessary imports for the given service.
func (t *TornadoGenerator) GenerateServiceImports(file *os.File, s *parser.Service) error {
	imports := "from datetime import timedelta\n"
	imports += "from threading import Lock\n\n"

	imports += "from frugal.exceptions import FApplicationException\n"
	imports += "from frugal.exceptions import FMessageSizeException\n"
	imports += "from frugal.exceptions import FRateLimitException\n"
	imports += "from frugal.exceptions import FTimeoutException\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.tornado.processor import FBaseProcessor\n"
	imports += "from frugal.tornado.processor import FProcessorFunction\n"
	imports += "from frugal.transport import TMemoryOutputBuffer\n"
	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"
	imports += "from tornado import gen\n"
	imports += "from tornado.concurrent import Future\n\n"

	imports += t.generateServiceExtendsImport(s)
	imports += t.generateServiceIncludeImports(s)

	_, err := file.WriteString(imports)
	return err

}

// GenerateScopeImports generates necessary imports for the given scope.
func (t *TornadoGenerator) GenerateScopeImports(file *os.File, s *parser.Scope) error {
	imports := "import sys\n"
	imports += "import traceback\n\n"

	imports += "from thrift.Thrift import TApplicationException\n"
	imports += "from thrift.Thrift import TMessageType\n"
	imports += "from thrift.Thrift import TType\n"
	imports += "from tornado import gen\n"
	imports += "from frugal.middleware import Method\n"
	imports += "from frugal.subscription import FSubscription\n"
	imports += "from frugal.transport import TMemoryOutputBuffer\n\n"

	imports += "from .ttypes import *\n"
	_, err := file.WriteString(imports)
	return err
}

// GenerateService generates the given service.
func (t *TornadoGenerator) GenerateService(file *os.File, s *parser.Service) error {
	contents := ""
	contents += t.generateServiceInterface(s)
	contents += t.generateClient(s)
	contents += t.generateServer(s)
	contents += t.generateServiceArgsResults(s)

	_, err := file.WriteString(contents)
	return err
}

func (t *TornadoGenerator) generateClient(service *parser.Service) string {
	contents := "\n"
	contents += t.generateClientConstructor(service, true)
	for _, method := range service.Methods {
		contents += t.generateClientMethod(method)
	}
	contents += "\n"
	return contents
}

func (t *TornadoGenerator) generateClientMethod(method *parser.Method) string {
	contents := ""
	contents += t.generateMethodSignature(method)
	contents += tabtab + fmt.Sprintf("return self._methods['%s']([ctx%s])\n\n",
		method.Name, t.generateClientArgs(method.Arguments))

	contents += tab + "@gen.coroutine\n"
	contents += tab + fmt.Sprintf("def _%s(self, ctx%s):\n", method.Name, t.generateClientArgs(method.Arguments))

	if method.Oneway {
		contents += tabtab + fmt.Sprintf("yield self._send_%s(ctx%s)\n\n", method.Name, t.generateClientArgs(method.Arguments))
		contents += t.generateClientSendMethod(method)
		return contents
	}

	contents += tabtab + "delta = timedelta(milliseconds=ctx.timeout)\n"
	contents += tabtab + "callback_future = Future()\n"
	contents += tabtab + "timeout_future = gen.with_timeout(delta, callback_future)\n"
	contents += tabtab + fmt.Sprintf("self._transport.register(ctx, self._recv_%s(ctx, callback_future))\n", method.Name)
	contents += tabtab + "try:\n"
	contents += tabtabtab + fmt.Sprintf("yield self._send_%s(ctx%s)\n", method.Name, t.generateClientArgs(method.Arguments))
	contents += tabtabtab + "result = yield timeout_future\n"
	contents += tabtab + "except gen.TimeoutError:\n"
	contents += fmt.Sprintf(tabtabtab+"raise FTimeoutException('%s timed out after {} milliseconds'.format(ctx.timeout))\n", method.Name)
	contents += tabtab + "finally:\n"
	contents += tabtabtab + "self._transport.unregister(ctx)\n"
	contents += tabtab + "raise gen.Return(result)\n\n"
	contents += t.generateClientSendMethod(method)
	contents += t.generateClientRecvMethod(method)

	return contents
}

func (t *TornadoGenerator) generateClientSendMethod(method *parser.Method) string {
	contents := ""
	contents += tab + "@gen.coroutine\n"
	contents += tab + fmt.Sprintf("def _send_%s(self, ctx%s):\n", method.Name, t.generateClientArgs(method.Arguments))
	contents += tabtab + "buffer = TMemoryOutputBuffer(self._transport.get_request_size_limit())\n"
	contents += tabtab + "oprot = self._protocol_factory.get_protocol(buffer)\n"
	contents += tabtab + "oprot.write_request_headers(ctx)\n"
	contents += tabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.CALL, 0)\n", method.Name)
	contents += tabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	for _, arg := range method.Arguments {
		contents += tabtab + fmt.Sprintf("args.%s = %s\n", arg.Name, arg.Name)
	}
	contents += tabtab + "args.write(oprot)\n"
	contents += tabtab + "oprot.writeMessageEnd()\n"
	contents += tabtab + "yield self._transport.send(buffer.getvalue())\n\n"

	return contents
}

func (t *TornadoGenerator) generateClientRecvMethod(method *parser.Method) string {
	contents := tab + fmt.Sprintf("def _recv_%s(self, ctx, future):\n", method.Name)
	contents += tabtab + fmt.Sprintf("def %s_callback(transport):\n", method.Name)
	contents += tabtabtab + "iprot = self._protocol_factory.get_protocol(transport)\n"
	contents += tabtabtab + "iprot.read_response_headers(ctx)\n"
	contents += tabtabtab + "_, mtype, _ = iprot.readMessageBegin()\n"
	contents += tabtabtab + "if mtype == TMessageType.EXCEPTION:\n"
	contents += tabtabtabtab + "x = TApplicationException()\n"
	contents += tabtabtabtab + "x.read(iprot)\n"
	contents += tabtabtabtab + "iprot.readMessageEnd()\n"
	contents += tabtabtabtab + "if x.type == FApplicationException.RESPONSE_TOO_LARGE:\n"
	contents += tabtabtabtabtab + "future.set_exception(FMessageSizeException.response(x.message))\n"
	contents += tabtabtabtabtab + "return\n"
	contents += tabtabtabtab + "if x.type == FApplicationException.RATE_LIMIT_EXCEEDED:\n"
	contents += tabtabtabtabtab + "future.set_exception(FRateLimitException(x.message))\n"
	contents += tabtabtabtabtab + "return\n"
	contents += tabtabtabtab + "future.set_exception(x)\n"
	contents += tabtabtabtab + "return\n"
	contents += tabtabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	contents += tabtabtab + "result.read(iprot)\n"
	contents += tabtabtab + "iprot.readMessageEnd()\n"
	for _, err := range method.Exceptions {
		contents += tabtabtab + fmt.Sprintf("if result.%s is not None:\n", err.Name)
		contents += tabtabtabtab + fmt.Sprintf("future.set_exception(result.%s)\n", err.Name)
		contents += tabtabtabtab + "return\n"
	}
	if method.ReturnType == nil {
		contents += tabtabtab + "future.set_result(None)\n"
	} else {
		contents += tabtabtab + "if result.success is not None:\n"
		contents += tabtabtabtab + "future.set_result(result.success)\n"
		contents += tabtabtabtab + "return\n"
		contents += tabtabtab + fmt.Sprintf(
			"x = TApplicationException(TApplicationException.MISSING_RESULT, \"%s failed: unknown result\")\n", method.Name)
		contents += tabtabtab + "future.set_exception(x)\n"
		contents += tabtabtab + "raise x\n"
	}
	contents += tabtab + fmt.Sprintf("return %s_callback\n\n", method.Name)

	return contents
}

func (t *TornadoGenerator) generateServer(service *parser.Service) string {
	contents := ""
	contents += t.generateProcessor(service)
	for _, method := range service.Methods {
		contents += t.generateProcessorFunction(method)
	}

	contents += t.generateWriteApplicationException()
	return contents
}

func (t *TornadoGenerator) generateProcessorFunction(method *parser.Method) string {
	contents := ""
	contents += fmt.Sprintf("class _%s(FProcessorFunction):\n\n", method.Name)
	contents += tab + "def __init__(self, handler, lock):\n"
	contents += tabtab + "self._handler = handler\n"
	contents += tabtab + "self._lock = lock\n\n"

	contents += tab + "@gen.coroutine\n"
	contents += tab + "def process(self, ctx, iprot, oprot):\n"
	contents += tabtab + fmt.Sprintf("args = %s_args()\n", method.Name)
	contents += tabtab + "args.read(iprot)\n"
	contents += tabtab + "iprot.readMessageEnd()\n"
	if !method.Oneway {
		contents += tabtab + fmt.Sprintf("result = %s_result()\n", method.Name)
	}
	contents += tabtab + "try:\n"
	if method.ReturnType == nil {
		contents += tabtabtab + fmt.Sprintf("yield gen.maybe_future(self._handler([ctx%s]))\n",
			t.generateServerArgs(method.Arguments))
	} else {
		contents += tabtabtab + fmt.Sprintf("result.success = yield gen.maybe_future(self._handler([ctx%s]))\n",
			t.generateServerArgs(method.Arguments))
	}
	for _, err := range method.Exceptions {
		contents += tabtab + fmt.Sprintf("except %s as %s:\n", t.qualifiedTypeName(err.Type), err.Name)
		contents += tabtabtab + fmt.Sprintf("result.%s = %s\n", err.Name, err.Name)
	}
	contents += tabtab + "except FRateLimitException as ex:\n"
	contents += tabtabtab + "with (yield self._lock.acquire()):\n"
	contents += tabtabtabtab +
		fmt.Sprintf("_write_application_exception(ctx, oprot, FApplicationException.RATE_LIMIT_EXCEEDED, \"%s\", ex.message)\n",
			method.Name)
	contents += tabtabtabtab + "return\n"
	contents += tabtab + "except Exception as e:\n"
	if !method.Oneway {
		contents += tabtabtab + "with (yield self._lock.acquire()):\n"
		contents += tabtabtabtab + fmt.Sprintf("e = _write_application_exception(ctx, oprot, TApplicationException.UNKNOWN, \"%s\", e.message)\n", method.Name)
	}
	contents += tabtabtab + "raise e\n"
	if !method.Oneway {
		contents += tabtab + "with (yield self._lock.acquire()):\n"
		contents += tabtabtab + "try:\n"
		contents += tabtabtabtab + "oprot.write_response_headers(ctx)\n"
		contents += tabtabtabtab + fmt.Sprintf("oprot.writeMessageBegin('%s', TMessageType.REPLY, 0)\n", method.Name)
		contents += tabtabtabtab + "result.write(oprot)\n"
		contents += tabtabtabtab + "oprot.writeMessageEnd()\n"
		contents += tabtabtabtab + "oprot.get_transport().flush()\n"
		contents += tabtabtab + "except FMessageSizeException as e:\n"
		contents += tabtabtabtab + fmt.Sprintf(
			"raise _write_application_exception(ctx, oprot, FApplicationException.RESPONSE_TOO_LARGE, \"%s\", e.message)\n", method.Name)
	}
	contents += "\n\n"

	return contents
}

// GenerateSubscriber generates the subscriber for the given scope.
func (t *TornadoGenerator) GenerateSubscriber(file *os.File, scope *parser.Scope) error {
	subscriber := ""
	subscriber += fmt.Sprintf("class %sSubscriber(object):\n", scope.Name)
	if scope.Comment != nil {
		subscriber += t.generateDocString(scope.Comment, tab)
	}
	subscriber += "\n"

	subscriber += tab + fmt.Sprintf("_DELIMITER = '%s'\n\n", globals.TopicDelimiter)

	subscriber += tab + "def __init__(self, provider, middleware=None):\n"
	subscriber += t.generateDocString([]string{
		fmt.Sprintf("Create a new %sSubscriber.\n", scope.Name),
		"Args:",
		tab + "provider: FScopeProvider",
		tab + "middleware: ServiceMiddleware or list of ServiceMiddleware",
	}, tabtab)
	subscriber += "\n"
	subscriber += tabtab + "if middleware and not isinstance(middleware, list):\n"
	subscriber += tabtabtab + "middleware = [middleware]\n"
	subscriber += tabtab + "self._middleware = middleware\n"
	subscriber += tabtab + "self._provider = provider\n\n"

	for _, op := range scope.Operations {
		subscriber += t.generateSubscribeMethod(scope, op)
		subscriber += "\n\n"
	}

	_, err := file.WriteString(subscriber)
	return err
}

func (t *TornadoGenerator) generateSubscribeMethod(scope *parser.Scope, op *parser.Operation) string {
	args := ""
	docstr := []string{}
	if len(scope.Prefix.Variables) > 0 {
		docstr = append(docstr, "Args:")
		prefix := ""
		for _, variable := range scope.Prefix.Variables {
			docstr = append(docstr, tab+fmt.Sprintf("%s: string", variable))
			args += prefix + variable
			prefix = ", "
		}
		args += ", "
	}
	docstr = append(docstr, tab+fmt.Sprintf("%s_handler: function which takes FContext and %s", op.Name, op.Type))
	if op.Comment != nil {
		docstr[0] = "\n" + tabtab + docstr[0]
		docstr = append(op.Comment, docstr...)
	}
	method := tab + "@gen.coroutine\n"
	method += tab + fmt.Sprintf("def subscribe_%s(self, %s%s_handler):\n", op.Name, args, op.Name)
	method += t.generateDocString(docstr, tabtab)
	method += "\n"

	method += tabtab + fmt.Sprintf("op = '%s'\n", op.Name)
	method += tabtab + fmt.Sprintf("prefix = %s\n", generatePrefixStringTemplate(scope))
	method += tabtab + fmt.Sprintf("topic = '{}%s{}{}'.format(prefix, self._DELIMITER, op)\n\n", scope.Name)

	method += tabtab + "transport, protocol_factory = self._provider.new_subscriber()\n"
	method += tabtab + fmt.Sprintf(
		"yield transport.subscribe(topic, self._recv_%s(protocol_factory, op, %s_handler))\n\n",
		op.Name, op.Name)

	method += tab + fmt.Sprintf("def _recv_%s(self, protocol_factory, op, handler):\n", op.Name)
	method += tabtab + "method = Method(handler, self._middleware)\n\n"

	method += tabtab + "def callback(transport):\n"
	method += tabtabtab + "iprot = protocol_factory.get_protocol(transport)\n"
	method += tabtabtab + "ctx = iprot.read_request_headers()\n"
	method += tabtabtab + "mname, _, _ = iprot.readMessageBegin()\n"
	method += tabtabtab + "if mname != op:\n"
	method += tabtabtabtab + "iprot.skip(TType.STRUCT)\n"
	method += tabtabtabtab + "iprot.readMessageEnd()\n"
	method += tabtabtabtab + "raise TApplicationException(TApplicationException.UNKNOWN_METHOD)\n"
	method += tabtabtab + fmt.Sprintf("req = %s()\n", op.Type.Name)
	method += tabtabtab + "req.read(iprot)\n"
	method += tabtabtab + "iprot.readMessageEnd()\n"
	method += tabtabtab + "try:\n"
	method += tabtabtabtab + "method([ctx, req])\n"
	method += tabtabtab + "except:\n"
	method += tabtabtabtab + "traceback.print_exc()\n"
	method += tabtabtabtab + "sys.exit(1)\n\n"

	method += tabtab + "return callback\n\n"

	return method
}

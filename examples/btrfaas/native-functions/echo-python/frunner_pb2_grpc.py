# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import frunner_pb2 as frunner__pb2


class FunctionRunnerStub(object):
  """FunctionRunner is a service capable of running a function via the Run() method
  the options for the function(s) are stored in the requests metadata under the key `options`
  The `frunner` implementation expects the following metadata structure:
  metadata = { options: [ "key1=val1", "key2=val2" ] }
  The `fgateway` implementation expects this metadata structure:
  metadata = {
  chain: [ "function1", "function2"],
  options: [ "json-string-string-map-1", "json-string-string-map-2"]
  }
  """

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Run = channel.stream_stream(
        '/grpc.FunctionRunner/Run',
        request_serializer=frunner__pb2.Data.SerializeToString,
        response_deserializer=frunner__pb2.Data.FromString,
        )


class FunctionRunnerServicer(object):
  """FunctionRunner is a service capable of running a function via the Run() method
  the options for the function(s) are stored in the requests metadata under the key `options`
  The `frunner` implementation expects the following metadata structure:
  metadata = { options: [ "key1=val1", "key2=val2" ] }
  The `fgateway` implementation expects this metadata structure:
  metadata = {
  chain: [ "function1", "function2"],
  options: [ "json-string-string-map-1", "json-string-string-map-2"]
  }
  """

  def Run(self, request_iterator, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_FunctionRunnerServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'Run': grpc.stream_stream_rpc_method_handler(
          servicer.Run,
          request_deserializer=frunner__pb2.Data.FromString,
          response_serializer=frunner__pb2.Data.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'grpc.FunctionRunner', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))
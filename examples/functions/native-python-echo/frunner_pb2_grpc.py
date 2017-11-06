# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import frunner_pb2 as frunner__pb2


class FunctionRunnerStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.Run = channel.stream_stream(
        '/grpc.FunctionRunner/Run',
        request_serializer=frunner__pb2.FrunnerInputData.SerializeToString,
        response_deserializer=frunner__pb2.FrunnerOutputData.FromString,
        )


class FunctionRunnerServicer(object):
  # missing associated documentation comment in .proto file
  pass

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
          request_deserializer=frunner__pb2.FrunnerInputData.FromString,
          response_serializer=frunner__pb2.FrunnerOutputData.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'grpc.FunctionRunner', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))

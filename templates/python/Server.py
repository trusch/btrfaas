# Copyright 2015 gRPC authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""The Python implementation of the GRPC frunner.Server server."""

from concurrent import futures
import time
import os

import grpc

import frunner_pb2
import frunner_pb2_grpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class Server(frunner_pb2_grpc.FunctionRunnerServicer):

    def Run(self, request_iterator, context):
        for inputData in request_iterator:
            output = frunner_pb2.Data()
            output.data = inputData.data
            yield output


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    frunner_pb2_grpc.add_FunctionRunnerServicer_to_server(Server(), server)
    caPath = "/run/secrets/btrfaas-ca-cert.pem"
    if os.path.isdir(caPath):
        caPath += "/value"
    keyPath = "/run/secrets/btrfaas-function-key.pem"
    if os.path.isdir(keyPath):
        keyPath += "/value"
    certPath = "/run/secrets/btrfaas-function-cert.pem"
    if os.path.isdir(certPath):
        certPath += "/value"
    caCertFile = open(caPath, "rb")
    certFile = open(certPath, "rb")
    keyFile = open(keyPath, "rb")
    caCert = caCertFile.read()
    cert = certFile.read()
    key = keyFile.read()
    creds = grpc.ssl_server_credentials(((key,cert),), root_certificates=caCert, require_client_auth=True)
    server.add_secure_port('[::]:2424', creds)
    server.start()
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS);
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()

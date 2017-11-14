
/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

var PROTO_PATH = __dirname + '/../../../protos/route_guide.proto';

var grpc = require('grpc');
var frunner = grpc.load('./frunner.proto').grpc;

function run(call) {
  call.on('data', function(inputData) {
    call.write({data: inputData.data});
  });
  call.on('end', function() {
    call.end();
  });
}

function getServer() {
  var server = new grpc.Server();
  server.addService(frunner.FunctionRunner.service, {
    run: run,
  });
  return server;
}

if (require.main === module) {
  // If this is run as a script, start a server on an unused port
  var functionRunner = getServer();
  functionRunner.bind('0.0.0.0:2424', grpc.ServerCredentials.createInsecure());
  functionRunner.start();
}

exports.getServer = getServer;

# Copyright 2017 Google Inc. All Rights Reserved.
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

# build
FROM golang:1.11.5 as builder
WORKDIR /go/src/FPSSample

RUN apt-get update && apt-get install -y unzip

COPY examples/FPSSample/main.go .
COPY . /go/src/agones.dev/agones
RUN go build -o wrapper .

COPY examples/FPSSample/FPSSampleServer.zip .
RUN unzip FPSSampleServer.zip && rm FPSSampleServer.zip

# final image
FROM ubuntu:18.04

WORKDIR /home/FPSSample
COPY --from=builder /go/src/FPSSample/wrapper .
COPY --from=builder /go/src/FPSSample/FPSSampleServer .
RUN chmod +x wrapper
COPY examples/FPSSample/startup.sh .

EXPOSE 7913

ENTRYPOINT ./wrapper -i ./startup.sh -r 'Restarting gamdemode'
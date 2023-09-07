FROM nvidia/cuda:12.2.0-devel-ubuntu22.04 as build

RUN apt-get update && apt-get install -y \
  git \
  build-essential \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /src

RUN git clone https://github.com/ggerganov/llama.cpp.git .
RUN make LLAMA_CUBLAS=1


FROM nvidia/cuda:12.2.0-runtime-ubuntu22.04

RUN apt-get update && apt-get install -y \
  tini \
  && rm -rf /var/lib/apt/lists/*

COPY --from=build /src/server .
COPY --from=build /src/examples/server/public .

ENTRYPOINT [ "tini", "--", "/server" ]
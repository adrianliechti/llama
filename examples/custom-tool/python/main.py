import json
import grpc

import tool_pb2
import tool_pb2_grpc

from concurrent import futures
from grpc_reflection.v1alpha import reflection

class ToolServicer(tool_pb2_grpc.ToolServicer):
    def Tools(self, request, context):
        return tool_pb2.ToolsResponse(
            definitions=[
                tool_pb2.Definition(
                    name="get_weather",
                    description="get the weather for a given location.",
                    parameters=json.dumps({
                        "type": "object",

                        "properties": {
                            "location": {
                                "type": "array",

                                "items": {
                                    "type": "string",
                                },
                            },
                        },

                        "required": ["location"],
                    }),
                ),
            ],
        )

    def Execute(self, request, context):
        print(request.name, request.parameters)

        params = json.loads(request.parameters)
        location = params["location"]
        
        return tool_pb2.ResultResponse(data=f"It is always sunny in {location}!!!")

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    tool_pb2_grpc.add_ToolServicer_to_server(ToolServicer(), server)

    SERVICE_NAMES = (
        tool_pb2.DESCRIPTOR.services_by_name['Tool'].full_name,
        reflection.SERVICE_NAME,
    )

    reflection.enable_server_reflection(SERVICE_NAMES, server)

    server.add_insecure_port('[::]:50051')
    server.start()

    print("Tool Server started. Listening on port 50051.")
    
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
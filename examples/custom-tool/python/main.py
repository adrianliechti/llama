import json
import grpc

import tool_pb2
import tool_pb2_grpc

from concurrent import futures

class ToolServicer(tool_pb2_grpc.ToolServicer):
    def Info(self, request, context):
        d = tool_pb2.Definition()
        
        d.name = "get_weather"
        d.description = "get the weather for a given location."
        d.schema = json.dumps({
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
        })

        return d
    
    def Execute(self, request, context):
        params = json.loads(request.parameter)
        location = params["location"]
        
        r = tool_pb2.Result()
        r.content = f"It is always sunny in {location}!!!"

        return r

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    tool_pb2_grpc.add_ToolServicer_to_server(ToolServicer(), server)

    server.add_insecure_port('[::]:50051')
    server.start()

    print("Tool Server started. Listening on port 50051.")
    
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
from flask import Flask, jsonify, Response
import os
import yaml

app = Flask(__name__)

@app.route('/')
def hello():
    format = os.getenv("RESPONSE_FORMAT", "text")
    language = os.getenv("RESPONSE_LANGUAGE", "python")
    ros = os.getenv("RESPONSE_OS", "ubuntu")

    # Set custom headers for inference
    response = jsonify(message="Hello, World!")
    response.headers['X-Server-Language'] = language
    response.headers['X-Server-OS'] = ros

    

    if format == "json":
        response.mimetype = "application/json"
        response.data = '{"message": "Hello, World!"}'
    elif format == "yaml":
        response.data = yaml.dump({"message": "Hello, World!"})
        response.mimetype = "application/x-yaml"
    else:
        response.data = "Hello, World!"
        response.mimetype = "text/plain"

    return response

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)

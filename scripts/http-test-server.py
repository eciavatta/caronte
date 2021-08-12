#!/usr/bin/env python3
"""
Usage:
    ./http-test-server.py [port]
"""
from http.server import BaseHTTPRequestHandler, HTTPServer
import logging
import json
import sys

class ServerHandler(BaseHTTPRequestHandler):
    def response(self, data, status_code=200):
        self.send_response(status_code)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(data).encode('utf-8'))

    def do_GET(self):
        if (self.path.startswith('/numbers')):
            return self.response(list(range(100)))

        self.response({"error": "not found"}, status_code=404)

def run(server_class=HTTPServer, handler_class=ServerHandler, port=8080):
    logging.basicConfig(level=logging.INFO)
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    logging.info('Starting server...')
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()
    logging.info('Stopping server...')

if __name__ == '__main__':
    if len(sys.argv) == 2:
        run(port=int(sys.argv[1]))
    else:
        run()

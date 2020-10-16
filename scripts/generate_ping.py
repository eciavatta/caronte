#!/usr/bin/env python3

import socket
import socketserver
import sys
import time

class PongHandler(socketserver.BaseRequestHandler):

    def handle(self):
        def recv_data():
            return self.request.recv(1024).strip()

        data = recv_data()
        while len(data) > 0:
            counter = str(int(data) + 1).rjust(4, "0")
            print(f"PONG! {counter}", end='\r')
            self.request.sendall(counter.encode())
            data = recv_data()


if __name__ == "__main__":
    if not ((sys.argv[1] == "server" and len(sys.argv) == 2) or
        (sys.argv[1] == "client" and len(sys.argv) == 3)):
        print(f"Usage: {sys.argv[0]} server")
        print(f"       {sys.argv[0]} client <server_address>")
        exit(1)

    port = 9999
    n = 10000

    if sys.argv[1] == "server":
        # docker run -it --rm -p 9999:9999 -v "$PWD":/ping -w /ping python:3 python generate_ping.py server
        with socketserver.TCPServer(("0.0.0.0", port), PongHandler) as server:
            server.serve_forever()
    else:
        # docker run -it --rm -v "$PWD":/ping -w /ping python:3 python generate_ping.py client <server_address>
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.connect((sys.argv[2], port))

            counter = 0
            while counter < n:
                counter = str(counter).rjust(4, "0")
                print(f"PING! {counter}", end='\r')
                time.sleep(0.05)
                sock.sendall(counter.encode())
                counter = int(sock.recv(1024).strip()) + 1


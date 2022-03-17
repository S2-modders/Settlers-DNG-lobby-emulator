#! /usr/bin/env python3

import subprocess
from http.server import HTTPServer, BaseHTTPRequestHandler

BASE_PORT = 10000

class S2Controller(BaseHTTPRequestHandler):

    def do_GET(self):
        print(self.path)

        if self.path == "/api/request":
            port = self._requestAvailablePort()
            print(port)

            self.send_response(200)
            self.send_header("Content-type", "text/plain")
            self.end_headers()

            self.wfile.write(str(port).encode())

        else:
            self.send_error(404, "the fuck are you going??")
            self.end_headers()

    def _requestAvailablePort(self) -> int:
        port = BASE_PORT
        while self.__checkPortAvailable(port) == False:
            port += 1

        return port

    def __checkPortAvailable(self, port: int) -> bool:
        command = f"bash portcheck.sh {port}"    
        res = subprocess.check_output(command.split()).strip().decode()

        return int(res) == 0

def main():
    with HTTPServer(("", 5480), S2Controller) as httpServer:
        httpServer.serve_forever()

if __name__ == "__main__":
    print("starting S2Controller")
    main()

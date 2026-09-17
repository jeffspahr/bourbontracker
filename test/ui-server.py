"""Serve the UI with repository fixtures: python3 test/ui-server.py [port]."""
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[1]


class FixtureHandler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(ROOT), **kwargs)

    def do_GET(self):
        fixtures = {
            '/inventory-va.json': '/test/new-inventory-va.json',
            '/inventory-nc.json': '/test/new-inventory-nc.json',
        }
        self.path = fixtures.get(self.path, self.path)
        super().do_GET()


if __name__ == '__main__':
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8001
    print(f'UI with fixture inventories: http://localhost:{port}', flush=True)
    ThreadingHTTPServer(('127.0.0.1', port), FixtureHandler).serve_forever()

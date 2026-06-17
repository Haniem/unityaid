import json
import ssl
import struct
import sys
import time
import urllib.parse
import urllib.request

import websocket
from Crypto.Cipher import DES

text = sys.argv[1] if len(sys.argv) > 1 else ""
if "\\n" in text:
    text = text.replace("\\n", "\n")

base = "https://ru213.vdska.ru:8006/api2/json"
ctx = ssl._create_unverified_context()


def post(path, data, cookie=None, csrf=None):
    body = urllib.parse.urlencode(data).encode()
    req = urllib.request.Request(base + path, data=body, method="POST")
    if cookie:
        req.add_header("Cookie", "PVEAuthCookie=" + cookie)
    if csrf:
        req.add_header("CSRFPreventionToken", csrf)
    with urllib.request.urlopen(req, context=ctx, timeout=20) as response:
        return json.loads(response.read())


def reverse_bits(byte):
    return int(f"{byte:08b}"[::-1], 2)


def rfb_password_response(challenge, password):
    key = (password.encode("latin1", "ignore")[:8]).ljust(8, b"\0")
    key = bytes(reverse_bits(byte) for byte in key)
    return DES.new(key, DES.MODE_ECB).encrypt(challenge)


def send_key(ws, keysym):
    ws.send_binary(struct.pack(">BBHI", 4, 1, 0, keysym))
    time.sleep(0.01)
    ws.send_binary(struct.pack(">BBHI", 4, 0, 0, keysym))
    time.sleep(0.01)


login = post(
    "/access/ticket",
    {"username": "haniemcs@gmail.com_163802@pve", "password": "LGWoXx5s9A3O"},
)["data"]
vnc = post(
    "/nodes/ru213/qemu/163802/vncproxy",
    {},
    login["ticket"],
    login["CSRFPreventionToken"],
)["data"]
url = (
    "wss://ru213.vdska.ru:8006/api2/json/nodes/ru213/qemu/163802/vncwebsocket"
    "?port="
    + urllib.parse.quote(str(vnc["port"]))
    + "&vncticket="
    + urllib.parse.quote(vnc["ticket"])
)
ws = websocket.create_connection(
    url,
    sslopt={"cert_reqs": ssl.CERT_NONE},
    header=[f'Cookie: PVEAuthCookie={login["ticket"]}'],
    timeout=20,
)
ws.recv()
ws.send_binary(b"RFB 003.008\n")
ws.recv()
ws.send_binary(b"\x02")
challenge = ws.recv()
ws.send_binary(rfb_password_response(challenge, vnc["ticket"]))
ws.recv()
ws.send_binary(b"\x01")
ws.recv()

for ch in text:
    if ch == "\n":
        send_key(ws, 0xFF0D)
    else:
        send_key(ws, ord(ch))

ws.close()
print("typed")

import urllib.request, urllib.parse, json, ssl, websocket, struct, sys
from Crypto.Cipher import DES
from PIL import Image
out=sys.argv[1] if len(sys.argv)>1 else 'vnc_screen.png'
base='https://ru213.vdska.ru:8006/api2/json'; ctx=ssl._create_unverified_context()
def post(path,data,cookie=None,csrf=None):
 body=urllib.parse.urlencode(data).encode(); req=urllib.request.Request(base+path,data=body,method='POST')
 if cookie: req.add_header('Cookie','PVEAuthCookie='+cookie)
 if csrf: req.add_header('CSRFPreventionToken',csrf)
 with urllib.request.urlopen(req,context=ctx,timeout=20) as r: return json.loads(r.read())
def rb(b): return int('{:08b}'.format(b)[::-1],2)
def rp(ch,p):
 key=(p.encode('latin1','ignore')[:8]).ljust(8,b'\0'); key=bytes(rb(b) for b in key); return DES.new(key,DES.MODE_ECB).encrypt(ch)
login=post('/access/ticket',{'username':'haniemcs@gmail.com_163802@pve','password':'LGWoXx5s9A3O'})['data']; vnc=post('/nodes/ru213/qemu/163802/vncproxy',{},login['ticket'],login['CSRFPreventionToken'])['data']
url='wss://ru213.vdska.ru:8006/api2/json/nodes/ru213/qemu/163802/vncwebsocket?port='+urllib.parse.quote(str(vnc['port']))+'&vncticket='+urllib.parse.quote(vnc['ticket'])
ws=websocket.create_connection(url,sslopt={'cert_reqs':ssl.CERT_NONE},header=[f'Cookie: PVEAuthCookie={login["ticket"]}'],timeout=20)
ws.recv(); ws.send_binary(b'RFB 003.008\n'); ws.recv(); ws.send_binary(b'\x02'); ch=ws.recv(); ws.send_binary(rp(ch,vnc['ticket'])); ws.recv(); ws.send_binary(b'\x01'); init=ws.recv(); w,h=struct.unpack('>HH',init[:4])
class R:
 def __init__(self,ws): self.ws=ws; self.buf=b''
 def read(self,n):
  while len(self.buf)<n: self.buf+=self.ws.recv()
  o=self.buf[:n]; self.buf=self.buf[n:]; return o
r=R(ws); ws.send_binary(b'\x00\x00\x00\x00'+struct.pack('>BBBBHHHBBBxxx',32,24,0,1,255,255,255,16,8,0)); ws.send_binary(struct.pack('>BBHi',2,0,1,0)); ws.send_binary(struct.pack('>BBHHHH',3,0,0,0,w,h))
while True:
 if r.read(1)[0]==0:
  r.read(1); n=struct.unpack('>H',r.read(2))[0]; img=Image.new('RGB',(w,h),'black')
  for i in range(n):
   x,y,rw,rh,enc=struct.unpack('>HHHHi',r.read(12)); raw=r.read(rw*rh*4); img.paste(Image.frombytes('RGB',(rw,rh),raw,'raw','BGRX'),(x,y))
  img.save(out); print(out); break
ws.close()

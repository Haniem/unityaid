from pathlib import Path
from PIL import Image, ImageDraw, ImageFont
import math
src=Path(r'E:/diploma/outputs/019ed18e-8ec4-7a51-b7bc-1a735bc7d652/presentations/vkr-defense/analysis')
imgs=[]
for p in sorted([*src.glob('image*.png'), *src.glob('image*.jpg')], key=lambda x: int(''.join(filter(str.isdigit,x.stem)) or 0)):
    try:
        im=Image.open(p); im.verify()
        im=Image.open(p).convert('RGB')
        if im.width<5 or im.height<5: continue
        imgs.append((p,im.copy()))
    except Exception: pass
thumb_w, thumb_h = 220, 140
cols=5; rows=math.ceil(len(imgs)/cols)
sheet=Image.new('RGB',(cols*thumb_w, rows*(thumb_h+28)), 'white')
d=ImageDraw.Draw(sheet)
for idx,(p,im) in enumerate(imgs):
    im.thumbnail((thumb_w-10, thumb_h-10))
    x=(idx%cols)*thumb_w+(thumb_w-im.width)//2; y=(idx//cols)*(thumb_h+28)+5
    sheet.paste(im,(x,y))
    d.text(((idx%cols)*thumb_w+8,(idx//cols)*(thumb_h+28)+thumb_h+8),p.name,fill=(0,0,0))
out=src/'media_contact_sheet.jpg'
sheet.save(out,quality=90)
print(out)

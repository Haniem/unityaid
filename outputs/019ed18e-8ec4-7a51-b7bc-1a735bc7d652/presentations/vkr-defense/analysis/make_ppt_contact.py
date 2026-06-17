from pathlib import Path
from PIL import Image, ImageDraw
import math, re
src=Path(r'E:/diploma/outputs/019ed18e-8ec4-7a51-b7bc-1a735bc7d652/presentations/vkr-defense/output/pptx_render')
files=sorted(src.glob('*.PNG'), key=lambda p:int(re.search(r'(\d+)', p.stem).group(1)))
thumb_w, thumb_h = 320, 180
cols=2; rows=math.ceil(len(files)/cols)
sheet=Image.new('RGB',(cols*thumb_w, rows*(thumb_h+26)), 'white')
d=ImageDraw.Draw(sheet)
for idx,p in enumerate(files):
    im=Image.open(p).convert('RGB'); im.thumbnail((thumb_w,thumb_h))
    x=(idx%cols)*thumb_w; y=(idx//cols)*(thumb_h+26)
    sheet.paste(im,(x,y))
    d.text((x+6,y+thumb_h+6),p.name,fill=(0,0,0))
out=src.parent/'contact_sheet.png'
sheet.save(out)
print(out)

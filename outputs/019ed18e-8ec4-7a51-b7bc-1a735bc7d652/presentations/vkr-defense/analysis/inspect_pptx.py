import zipfile, json, re
from pathlib import Path
from lxml import etree
pptx=Path(r'E:/diploma/outputs/019ed18e-8ec4-7a51-b7bc-1a735bc7d652/presentations/vkr-defense/template.pptx')
out=Path(r'E:/diploma/outputs/019ed18e-8ec4-7a51-b7bc-1a735bc7d652/presentations/vkr-defense/template-python-inspect')
out.mkdir(parents=True, exist_ok=True)
ns={'p':'http://schemas.openxmlformats.org/presentationml/2006/main','a':'http://schemas.openxmlformats.org/drawingml/2006/main','r':'http://schemas.openxmlformats.org/officeDocument/2006/relationships'}
slides=[]
with zipfile.ZipFile(pptx) as z:
    names=z.namelist()
    slide_names=sorted([n for n in names if re.match(r'ppt/slides/slide\d+\.xml$', n)], key=lambda n:int(re.search(r'(\d+)', n).group(1)))
    media=[n for n in names if n.startswith('ppt/media/')]
    masters=[n for n in names if n.startswith('ppt/slideMasters/') and n.endswith('.xml')]
    layouts=[n for n in names if n.startswith('ppt/slideLayouts/') and n.endswith('.xml')]
    theme=[n for n in names if n.startswith('ppt/theme/')]
    for sn in slide_names:
        root=etree.fromstring(z.read(sn))
        texts=[]; shapes=[]
        for sp in root.xpath('.//p:sp', namespaces=ns):
            name=''; ph=''; tx=''
            cnv=sp.find('.//p:cNvPr', ns)
            if cnv is not None: name=cnv.get('name','')
            ph_el=sp.find('.//p:ph', ns)
            if ph_el is not None: ph=ph_el.get('type','placeholder')
            tx=' '.join([t.text for t in sp.xpath('.//a:t', namespaces=ns) if t.text]).strip()
            shapes.append({'name':name,'placeholder':ph,'text':tx[:500]})
            if tx: texts.append(tx)
        slides.append({'slide':len(slides)+1,'file':sn,'texts':texts,'shapes':shapes})
(out/'template_inventory.json').write_text(json.dumps({'slides':slides,'media':media,'masters':masters,'layouts':layouts,'theme':theme}, ensure_ascii=False, indent=2), encoding='utf-8')
print(json.dumps({'slides':len(slides),'media':len(media),'masters':len(masters),'layouts':len(layouts),'out':str(out)}, ensure_ascii=False, indent=2))

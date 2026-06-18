from pathlib import Path
import re
from reportlab.lib import colors
from reportlab.lib.enums import TA_CENTER, TA_LEFT
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import mm
from reportlab.platypus import (
    BaseDocTemplate, Frame, PageTemplate, Paragraph, Spacer, PageBreak,
    KeepTogether, Table, TableStyle, HRFlowable, ListFlowable, ListItem
)
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.lib.colors import HexColor

ROOT = Path('/Users/pavel/Desktop/Университет/unityaid')
MD = ROOT / 'docs/final/Zozin_P_A_VKR_Spech.md'
OUT = ROOT / 'docs/final/Zozin_P_A_VKR_Spech_UTF8.pdf'
FONT_DIR = ROOT / 'unityaid-back/internal/modules/certificates/assets'

pdfmetrics.registerFont(TTFont('DejaVu', str(FONT_DIR / 'DejaVuSans.ttf')))
pdfmetrics.registerFont(TTFont('DejaVu-Bold', str(FONT_DIR / 'DejaVuSans-Bold.ttf')))

PAGE_W, PAGE_H = A4
BLUE = HexColor('#2d3f91')
BLUE_DARK = HexColor('#223174')
RED = HexColor('#d73b31')
LIGHT_BLUE = HexColor('#eef2ff')
SOFT = HexColor('#f7f8fc')
TEXT = HexColor('#1f2433')
MUTED = HexColor('#5c6378')
BORDER = HexColor('#d8deef')

TITLE = 'Текст выступления и план демонстрации'
SUBTITLE = 'ВКР: информационная система управления волонтерами «Пульс»'
AUTHOR = 'Зозин Павел Алексеевич'

styles = getSampleStyleSheet()
styles.add(ParagraphStyle(
    name='DocTitle', fontName='DejaVu-Bold', fontSize=24, leading=30,
    textColor=BLUE_DARK, alignment=TA_CENTER, spaceAfter=8
))
styles.add(ParagraphStyle(
    name='DocSubtitle', fontName='DejaVu', fontSize=11, leading=16,
    textColor=MUTED, alignment=TA_CENTER, spaceAfter=18
))
styles.add(ParagraphStyle(
    name='Section', fontName='DejaVu-Bold', fontSize=17, leading=22,
    textColor=colors.white, alignment=TA_LEFT, leftIndent=0, spaceBefore=10,
    spaceAfter=8
))
styles.add(ParagraphStyle(
    name='SlideTitle', fontName='DejaVu-Bold', fontSize=13.5, leading=17,
    textColor=BLUE_DARK, spaceBefore=7, spaceAfter=5, keepWithNext=True
))
styles.add(ParagraphStyle(
    name='StepTitle', fontName='DejaVu-Bold', fontSize=12.5, leading=16,
    textColor=BLUE_DARK, spaceBefore=8, spaceAfter=4, keepWithNext=True
))
styles.add(ParagraphStyle(
    name='BodyRu', fontName='DejaVu', fontSize=9.8, leading=14.2,
    textColor=TEXT, spaceAfter=5.2, firstLineIndent=0
))
styles.add(ParagraphStyle(
    name='BodyTight', fontName='DejaVu', fontSize=9.2, leading=13.0,
    textColor=TEXT, spaceAfter=3.5
))
styles.add(ParagraphStyle(
    name='Label', fontName='DejaVu-Bold', fontSize=9.5, leading=12,
    textColor=BLUE_DARK, spaceBefore=3, spaceAfter=2
))
styles.add(ParagraphStyle(
    name='CodeRu', fontName='DejaVu-Bold', fontSize=9.2, leading=12,
    textColor=HexColor('#0f5132'), backColor=HexColor('#eef8f2'), borderPadding=3,
    spaceAfter=3
))
styles.add(ParagraphStyle(
    name='Small', fontName='DejaVu', fontSize=8, leading=10,
    textColor=MUTED
))
styles.add(ParagraphStyle(
    name='BulletRu', parent=styles['BodyTight'], leftIndent=0, firstLineIndent=0
))


def esc(s: str) -> str:
    return (s.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;'))


def inline(s: str) -> str:
    s = esc(s.strip())
    s = re.sub(r'`([^`]+)`', r'<font name="DejaVu-Bold" color="#0f5132">\1</font>', s)
    s = s.replace(' - ', ' - ')
    return s


def read_sections():
    lines = MD.read_text(encoding='utf-8').splitlines()
    sections = []
    current_h2 = None
    current_h3 = None
    for raw in lines:
        line = raw.rstrip()
        if not line:
            continue
        if line.startswith('# '):
            continue
        if line.startswith('## '):
            current_h2 = {'title': line[3:].strip(), 'items': []}
            sections.append(current_h2)
            current_h3 = None
            continue
        if line.startswith('### '):
            current_h3 = {'title': line[4:].strip(), 'paras': []}
            if current_h2 is None:
                current_h2 = {'title': '', 'items': []}
                sections.append(current_h2)
            current_h2['items'].append(current_h3)
            continue
        if current_h3 is None:
            current_h3 = {'title': None, 'paras': []}
            if current_h2 is None:
                current_h2 = {'title': '', 'items': []}
                sections.append(current_h2)
            current_h2['items'].append(current_h3)
        current_h3['paras'].append(line)
    return sections


def section_band(title):
    p = Paragraph(inline(title), styles['Section'])
    tbl = Table([[p]], colWidths=[170*mm])
    tbl.setStyle(TableStyle([
        ('BACKGROUND', (0, 0), (-1, -1), BLUE),
        ('BOX', (0, 0), (-1, -1), 0, BLUE),
        ('LEFTPADDING', (0, 0), (-1, -1), 9),
        ('RIGHTPADDING', (0, 0), (-1, -1), 9),
        ('TOPPADDING', (0, 0), (-1, -1), 6),
        ('BOTTOMPADDING', (0, 0), (-1, -1), 6),
    ]))
    return tbl


def callout(title, body_lines):
    content = []
    if title:
        style = styles['SlideTitle'] if title.startswith('Слайд') or title.startswith('Слайды') else styles['StepTitle']
        content.append(Paragraph(inline(title), style))
    bullets = []
    pending_label = None
    for line in body_lines:
        if line.startswith('- '):
            bullets.append(ListItem(Paragraph(inline(line[2:]), styles['BodyTight']), leftIndent=8))
            continue
        if bullets:
            content.append(ListFlowable(bullets, bulletType='bullet', leftIndent=14, bulletFontName='DejaVu', bulletFontSize=7))
            bullets = []
        if line.endswith(':') and len(line) < 45:
            content.append(Paragraph(inline(line), styles['Label']))
        elif line.startswith('«') or line.startswith('"'):
            content.append(Paragraph(inline(line), styles['BodyRu']))
        else:
            content.append(Paragraph(inline(line), styles['BodyRu']))
    if bullets:
        content.append(ListFlowable(bullets, bulletType='bullet', leftIndent=14, bulletFontName='DejaVu', bulletFontSize=7))

    inner = Table([[content]], colWidths=[164*mm])
    inner.setStyle(TableStyle([
        ('BACKGROUND', (0, 0), (-1, -1), colors.white),
        ('BOX', (0, 0), (-1, -1), 0.65, BORDER),
        ('LEFTPADDING', (0, 0), (-1, -1), 9),
        ('RIGHTPADDING', (0, 0), (-1, -1), 9),
        ('TOPPADDING', (0, 0), (-1, -1), 6),
        ('BOTTOMPADDING', (0, 0), (-1, -1), 6),
    ]))
    return KeepTogether([inner, Spacer(1, 4)])


def cover_block():
    data = [[
        Paragraph('<font color="white">Пульс</font>', ParagraphStyle('Badge', fontName='DejaVu-Bold', fontSize=15, leading=18, alignment=TA_CENTER)),
        Paragraph(TITLE, styles['DocTitle'])
    ], [
        '', Paragraph(SUBTITLE + '<br/><br/>' + AUTHOR, styles['DocSubtitle'])
    ]]
    tbl = Table(data, colWidths=[35*mm, 135*mm])
    tbl.setStyle(TableStyle([
        ('BACKGROUND', (0, 0), (0, 1), BLUE),
        ('VALIGN', (0, 0), (-1, -1), 'MIDDLE'),
        ('SPAN', (0, 0), (0, 1)),
        ('LEFTPADDING', (0, 0), (-1, -1), 8),
        ('RIGHTPADDING', (0, 0), (-1, -1), 8),
        ('TOPPADDING', (0, 0), (-1, -1), 12),
        ('BOTTOMPADDING', (0, 0), (-1, -1), 12),
        ('BOX', (0, 0), (-1, -1), 0.8, BORDER),
        ('BACKGROUND', (1, 0), (1, 1), SOFT),
    ]))
    return tbl


class SpeechDoc(BaseDocTemplate):
    def __init__(self, filename):
        super().__init__(filename, pagesize=A4,
                         leftMargin=18*mm, rightMargin=18*mm,
                         topMargin=17*mm, bottomMargin=16*mm)
        frame = Frame(self.leftMargin, self.bottomMargin, self.width, self.height, id='normal')
        self.addPageTemplates([PageTemplate(id='speech', frames=[frame], onPage=self.decorate)])

    def decorate(self, canvas, doc):
        canvas.saveState()
        canvas.setFillColor(BLUE)
        canvas.rect(0, PAGE_H - 7*mm, PAGE_W, 7*mm, stroke=0, fill=1)
        canvas.setFillColor(RED)
        canvas.rect(0, PAGE_H - 7*mm, 20*mm, 7*mm, stroke=0, fill=1)
        canvas.setStrokeColor(BORDER)
        canvas.setLineWidth(0.4)
        canvas.line(18*mm, 13*mm, PAGE_W - 18*mm, 13*mm)
        canvas.setFont('DejaVu', 7.5)
        canvas.setFillColor(MUTED)
        canvas.drawString(18*mm, 8*mm, 'Текст выступления ВКР - информационная система «Пульс»')
        canvas.drawRightString(PAGE_W - 18*mm, 8*mm, f'{doc.page}')
        canvas.restoreState()


sections = read_sections()
story = [cover_block(), Spacer(1, 10), HRFlowable(width='100%', thickness=1, color=BORDER), Spacer(1, 6)]

first_section = True
for sec in sections:
    title = sec['title']
    if title:
        if not first_section:
            story.append(Spacer(1, 6))
        story.append(section_band(title))
        story.append(Spacer(1, 5))
        first_section = False
    for item in sec['items']:
        story.append(callout(item['title'], item['paras']))

OUT.parent.mkdir(parents=True, exist_ok=True)
doc = SpeechDoc(str(OUT))
doc.build(story)
print(OUT)

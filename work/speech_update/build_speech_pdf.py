from pathlib import Path
import re

from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import ParagraphStyle, getSampleStyleSheet
from reportlab.lib.units import cm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import (
    ListFlowable,
    ListItem,
    PageBreak,
    Paragraph,
    SimpleDocTemplate,
    Spacer,
)
from reportlab.platypus.tableofcontents import TableOfContents


ROOT = Path(r"E:\diploma")
SOURCE = ROOT / "docs" / "final" / "Zozin_P_A_VKR_Speech_10min.md"
OUTPUT = ROOT / "docs" / "final" / "Zozin_P_A_VKR_Speech_10min.pdf"


def esc(text: str) -> str:
    return (
        text.replace("&", "&amp;")
        .replace("<", "&lt;")
        .replace(">", "&gt;")
        .replace(" - ", " - ")
    )


def inline(text: str) -> str:
    text = esc(text)
    text = re.sub(r"`([^`]+)`", r'<font face="ArialMono">\1</font>', text)
    text = re.sub(r'"([^"]+)"', r'"\1"', text)
    return text


class NumberedCanvasDoc(SimpleDocTemplate):
    pass


def footer(canvas, doc):
    canvas.saveState()
    canvas.setFont("Arial", 9)
    canvas.setFillColor(colors.HexColor("#666666"))
    canvas.drawRightString(A4[0] - 1.6 * cm, 1.1 * cm, str(canvas.getPageNumber()))
    canvas.restoreState()


def build():
    pdfmetrics.registerFont(TTFont("Arial", r"C:\Windows\Fonts\arial.ttf"))
    pdfmetrics.registerFont(TTFont("ArialBold", r"C:\Windows\Fonts\arialbd.ttf"))
    pdfmetrics.registerFont(TTFont("ArialItalic", r"C:\Windows\Fonts\ariali.ttf"))
    pdfmetrics.registerFont(TTFont("ArialMono", r"C:\Windows\Fonts\consola.ttf"))

    styles = getSampleStyleSheet()
    styles.add(
        ParagraphStyle(
            name="TitleRu",
            fontName="ArialBold",
            fontSize=18,
            leading=22,
            spaceAfter=10,
            textColor=colors.HexColor("#111111"),
        )
    )
    styles.add(
        ParagraphStyle(
            name="H1Ru",
            fontName="ArialBold",
            fontSize=15,
            leading=18,
            spaceBefore=14,
            spaceAfter=8,
            textColor=colors.HexColor("#1F4D78"),
        )
    )
    styles.add(
        ParagraphStyle(
            name="H2Ru",
            fontName="ArialBold",
            fontSize=12.5,
            leading=15,
            spaceBefore=10,
            spaceAfter=5,
            textColor=colors.HexColor("#2E74B5"),
        )
    )
    styles.add(
        ParagraphStyle(
            name="BodyRu",
            fontName="Arial",
            fontSize=10.5,
            leading=14,
            spaceAfter=6,
        )
    )
    styles.add(
        ParagraphStyle(
            name="NoteRu",
            fontName="ArialItalic",
            fontSize=9.5,
            leading=13,
            spaceAfter=8,
            textColor=colors.HexColor("#555555"),
        )
    )
    styles.add(
        ParagraphStyle(
            name="BulletRu",
            fontName="Arial",
            fontSize=10.5,
            leading=14,
            leftIndent=10,
            spaceAfter=3,
        )
    )

    story = []
    pending_bullets = []

    def flush_bullets():
        nonlocal pending_bullets
        if pending_bullets:
            story.append(
                ListFlowable(
                    [ListItem(Paragraph(inline(item), styles["BulletRu"])) for item in pending_bullets],
                    bulletType="bullet",
                    start="circle",
                    leftIndent=18,
                    bulletFontName="Arial",
                    bulletFontSize=8,
                )
            )
            story.append(Spacer(1, 4))
            pending_bullets = []

    for raw in SOURCE.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line:
            flush_bullets()
            continue

        if line.startswith("# "):
            flush_bullets()
            story.append(Paragraph(inline(line[2:]), styles["TitleRu"]))
            continue
        if line.startswith("## "):
            flush_bullets()
            if story:
                story.append(Spacer(1, 4))
            story.append(Paragraph(inline(line[3:]), styles["H1Ru"]))
            continue
        if line.startswith("### "):
            flush_bullets()
            story.append(Paragraph(inline(line[4:]), styles["H2Ru"]))
            continue
        if line.startswith("- "):
            pending_bullets.append(line[2:])
            continue
        if re.match(r"^\d+\. ", line):
            flush_bullets()
            story.append(Paragraph(inline(line), styles["BodyRu"]))
            continue

        flush_bullets()
        style = styles["NoteRu"] if line.startswith("Ориентир по времени:") else styles["BodyRu"]
        story.append(Paragraph(inline(line), style))

    flush_bullets()

    doc = NumberedCanvasDoc(
        str(OUTPUT),
        pagesize=A4,
        rightMargin=1.7 * cm,
        leftMargin=1.7 * cm,
        topMargin=1.6 * cm,
        bottomMargin=1.6 * cm,
        title="Текст выступления и план демонстрации",
        author="Зозин П.А.",
    )
    doc.build(story, onFirstPage=footer, onLaterPages=footer)
    print(OUTPUT)


if __name__ == "__main__":
    build()

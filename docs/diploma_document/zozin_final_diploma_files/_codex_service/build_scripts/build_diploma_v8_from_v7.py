from __future__ import annotations

import io
import re
import shutil
import xml.etree.ElementTree as ET
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile

from PIL import Image


FINAL_DIR = Path(r"E:\diploma\docs\diploma_document\zozin_final_diploma_files")
INPUT_DOCX = FINAL_DIR / "Zozin_v7.docx"
OUTPUT_DOCX = FINAL_DIR / "Zozin_v8.docx"
WORK_DIR = FINAL_DIR / "_codex_service" / "generated" / "V8_image_processing"

NS = {
    "a": "http://schemas.openxmlformats.org/drawingml/2006/main",
    "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
    "wp": "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing",
    "rel": "http://schemas.openxmlformats.org/package/2006/relationships",
}


def is_diagram_like(name: str, image: Image.Image) -> bool:
    width, height = image.size
    suffix = Path(name).suffix.lower()
    if suffix in {".jpg", ".jpeg"}:
        return False
    if width == 1400 and height in {850, 1120}:
        return True
    if width >= 1800:
        return True
    return False


def non_white_bbox(image: Image.Image, threshold: int = 245) -> tuple[int, int, int, int] | None:
    rgb = image.convert("RGB")
    min_x = rgb.width
    min_y = rgb.height
    max_x = -1
    max_y = -1
    px = rgb.load()
    for y in range(rgb.height):
        for x in range(rgb.width):
            r, g, b = px[x, y]
            if r < threshold or g < threshold or b < threshold:
                min_x = min(min_x, x)
                min_y = min(min_y, y)
                max_x = max(max_x, x)
                max_y = max(max_y, y)
    if max_x < 0:
        return None
    return min_x, min_y, max_x + 1, max_y + 1


def active_row_groups(image: Image.Image, threshold: int = 245) -> list[tuple[int, int]]:
    rgb = image.convert("RGB")
    px = rgb.load()
    row_limit = max(8, int(rgb.width * 0.002))
    active_rows: list[int] = []
    for y in range(rgb.height):
        count = 0
        for x in range(rgb.width):
            r, g, b = px[x, y]
            if r < threshold or g < threshold or b < threshold:
                count += 1
        if count > row_limit:
            active_rows.append(y)

    if not active_rows:
        return []

    groups: list[tuple[int, int]] = []
    start = prev = active_rows[0]
    for y in active_rows[1:]:
        if y - prev <= 8:
            prev = y
            continue
        groups.append((start, prev))
        start = prev = y
    groups.append((start, prev))
    return groups


def crop_diagram(image: Image.Image) -> tuple[Image.Image, bool]:
    rgb = image.convert("RGB")
    bbox = non_white_bbox(rgb)
    if not bbox:
        return image, False

    left, top, right, bottom = bbox
    groups = active_row_groups(rgb)

    crop_top = max(top - 22, 0)
    if len(groups) >= 2:
        first, second = groups[0], groups[1]
        has_separate_title = first[0] < 100 and second[0] - first[1] > 24
        if has_separate_title:
            crop_top = max(second[0] - 24, 0)

    crop_left = max(left - 28, 0)
    crop_right = min(right + 28, rgb.width)
    crop_bottom = min(bottom + 30, rgb.height)

    new_width = crop_right - crop_left
    new_height = crop_bottom - crop_top
    if new_width < 200 or new_height < 120:
        return image, False
    if new_width < rgb.width * 0.55 or new_height < rgb.height * 0.45:
        return image, False

    same_box = (
        crop_top == 0
        and crop_left == 0
        and crop_right == rgb.width
        and crop_bottom == rgb.height
    )
    if same_box:
        return image, False

    return rgb.crop((crop_left, crop_top, crop_right, crop_bottom)), True


def crop_media_entries(raw_entries: dict[str, bytes]) -> dict[str, tuple[bytes, tuple[int, int]]]:
    cropped: dict[str, tuple[bytes, tuple[int, int]]] = {}
    WORK_DIR.mkdir(parents=True, exist_ok=True)

    for name, data in raw_entries.items():
        if not name.startswith("word/media/"):
            continue
        try:
            image = Image.open(io.BytesIO(data))
            image.load()
        except Exception:
            continue
        if not is_diagram_like(name, image):
            continue

        new_image, changed = crop_diagram(image)
        if not changed:
            continue

        out = io.BytesIO()
        if Path(name).suffix.lower() in {".jpg", ".jpeg"}:
            new_image.save(out, format="JPEG", quality=95)
        else:
            new_image.save(out, format="PNG")
        cropped[name] = (out.getvalue(), new_image.size)
        new_image.save(WORK_DIR / Path(name).name)

    return cropped


def rels_for_document(raw_entries: dict[str, bytes]) -> dict[str, str]:
    rels_name = "word/_rels/document.xml.rels"
    if rels_name not in raw_entries:
        return {}
    root = ET.fromstring(raw_entries[rels_name])
    rels: dict[str, str] = {}
    for rel in root.findall("rel:Relationship", NS):
        rid = rel.attrib.get("Id")
        target = rel.attrib.get("Target")
        if rid and target and target.startswith("media/"):
            rels[rid] = "word/" + target
    return rels


def update_image_extents(xml_bytes: bytes, rels: dict[str, str], cropped_sizes: dict[str, tuple[int, int]]) -> bytes:
    try:
        root = ET.fromstring(xml_bytes)
    except ET.ParseError:
        return xml_bytes

    changed = False
    for node_name in ("inline", "anchor"):
        for drawing in root.findall(f".//wp:{node_name}", NS):
            blip = drawing.find(".//a:blip", NS)
            extent = drawing.find("wp:extent", NS)
            a_ext = drawing.find(".//a:xfrm/a:ext", NS)
            if blip is None or extent is None:
                continue
            rid = blip.attrib.get(f"{{{NS['r']}}}embed")
            media = rels.get(rid or "")
            if not media or media not in cropped_sizes:
                continue

            width_px, height_px = cropped_sizes[media]
            old_cx = int(extent.attrib.get("cx", "0"))
            if old_cx <= 0 or width_px <= 0:
                continue
            new_cy = int(old_cx * height_px / width_px)
            extent.attrib["cy"] = str(new_cy)
            if a_ext is not None:
                a_ext.attrib["cy"] = str(new_cy)
            changed = True

    if not changed:
        return xml_bytes
    return ET.tostring(root, encoding="utf-8", xml_declaration=True)


def clean_xml_text(xml_bytes: bytes, stats: dict[str, int]) -> bytes:
    text = xml_bytes.decode("utf-8", errors="ignore")

    for dash in ("—", "–", "−", "‑"):
        count = text.count(dash)
        if count:
            stats["dash_replacements"] += count
            text = text.replace(dash, "-")

    phrase_replacements = {
        "Раздел оформлен по структуре, использованной в примере Белобородова: для каждого экрана указано его назначение, приведен экранный снимок и перечислены основные функции, доступные пользователю. ": "",
        "Такой формат позволяет связать проверку работоспособности не только с маршрутом API, но и с фактическим пользовательским интерфейсом.": "Такой порядок описания показывает, какие задачи пользователь решает на каждом экране.",
        "В данном разделе": "В разделе",
        "Следует отметить, что ": "",
        "В рамках данной выпускной квалификационной работы": "В выпускной квалификационной работе",
        "В рамках данной работы": "В работе",
        "В рамках системы": "В системе",
    }

    for old, new in phrase_replacements.items():
        count = text.count(old)
        if count:
            stats["phrase_replacements"] += count
            text = text.replace(old, new)

    thus_count = text.count("Таким образом, ")
    if thus_count:
        stats["phrase_replacements"] += thus_count
        text = text.replace("Таким образом, ", "")

    sentence_starts = {
        ">ключевым отличием Пульс является": ">Ключевым отличием Пульс является",
        ">постановка цели переводит": ">Постановка цели переводит",
        ">разработка Пульс обусловлена": ">Разработка Пульс обусловлена",
        ">проектные результаты второй главы": ">Проектные результаты второй главы",
        ">основные экраны приложения подтверждают": ">Основные экраны приложения подтверждают",
        ">третья глава отделяет": ">Третья глава отделяет",
        ">алгоритмы обработки заявок": ">Алгоритмы обработки заявок",
        ">подтверждаемая данными динамика": ">Подтверждаемая данными динамика",
    }
    for old, new in sentence_starts.items():
        count = text.count(old)
        if count:
            stats["phrase_replacements"] += count
            text = text.replace(old, new)

    text = re.sub(r"[ \t]{2,}", " ", text)
    return text.encode("utf-8")


def main() -> None:
    if not INPUT_DOCX.exists():
        raise FileNotFoundError(INPUT_DOCX)

    if WORK_DIR.exists():
        shutil.rmtree(WORK_DIR)

    with ZipFile(INPUT_DOCX, "r") as zin:
        raw_entries = {name: zin.read(name) for name in zin.namelist()}

    cropped = crop_media_entries(raw_entries)
    cropped_sizes = {name: size for name, (_, size) in cropped.items()}
    rels = rels_for_document(raw_entries)
    stats = {"dash_replacements": 0, "phrase_replacements": 0}

    with ZipFile(OUTPUT_DOCX, "w", compression=ZIP_DEFLATED) as zout:
        for name, data in raw_entries.items():
            if name in cropped:
                data = cropped[name][0]
            elif name.endswith(".xml") and (name.startswith("word/") or name.startswith("docProps/")):
                data = clean_xml_text(data, stats)
                if name == "word/document.xml":
                    data = update_image_extents(data, rels, cropped_sizes)
            zout.writestr(name, data)

    print(f"created={OUTPUT_DOCX}")
    print(f"cropped_images={len(cropped)}")
    print(f"dash_replacements={stats['dash_replacements']}")
    print(f"phrase_replacements={stats['phrase_replacements']}")
    for name, (_, size) in sorted(cropped.items()):
        print(f"cropped {Path(name).name}: {size[0]}x{size[1]}")


if __name__ == "__main__":
    main()

from __future__ import annotations

import re
from pathlib import Path

from docx import Document
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.shared import Cm, Pt


FINAL_DIR = Path(r"E:\diploma\docs\diploma_document\zozin_final_diploma_files")
INPUT_DOCX = FINAL_DIR / "Zozin_v6.docx"
OUTPUT_DOCX = FINAL_DIR / "Zozin_v7.docx"
SCREENSHOT_DIR = FINAL_DIR / "_codex_service" / "generated" / "V5_generated" / "screenshots"


def iter_nonempty_paragraphs(doc: Document):
    for idx, paragraph in enumerate(doc.paragraphs):
        text = " ".join(paragraph.text.split())
        if text:
            yield idx, paragraph, text


def move_to_before(paragraph, target):
    body = target._p.getparent()
    body.remove(paragraph._p)
    body.insert(body.index(target._p), paragraph._p)


def add_before(doc: Document, target, text: str = "", style: str | None = None):
    paragraph = doc.add_paragraph(text, style=style)
    move_to_before(paragraph, target)
    return paragraph


def add_picture_before(doc: Document, target, image_path: Path, caption: str):
    paragraph = doc.add_paragraph()
    paragraph.alignment = WD_ALIGN_PARAGRAPH.CENTER
    run = paragraph.add_run()
    run.add_picture(str(image_path), width=Cm(15.5))
    move_to_before(paragraph, target)

    caption_paragraph = doc.add_paragraph(caption)
    caption_paragraph.alignment = WD_ALIGN_PARAGRAPH.CENTER
    move_to_before(caption_paragraph, target)


def add_bullets_before(doc: Document, target, items: list[str]):
    for item in items:
        paragraph = doc.add_paragraph(f"- {item}")
        paragraph.paragraph_format.left_indent = Cm(1.0)
        paragraph.paragraph_format.first_line_indent = Cm(-0.4)
        paragraph.paragraph_format.space_after = Pt(0)
        move_to_before(paragraph, target)


def replace_first(text: str, old: str, new: str) -> str:
    return text.replace(old, new, 1) if old in text else text


def renumber_chapter_three(doc: Document):
    replacements = [
        ("3.6 Выводы по третьему разделу", "3.7 Выводы по третьему разделу"),
        ("3.5.6 Рекомендации по развитию проверки и внедрению", "3.6.6 Рекомендации по развитию проверки и внедрению"),
        ("3.5.5 Регламент пилотного запуска и фиксации результатов", "3.6.5 Регламент пилотного запуска и фиксации результатов"),
        ("3.5.4 Метрики эксплуатации и ограничения результатов", "3.6.4 Метрики эксплуатации и ограничения результатов"),
        ("3.5.3 Контроль данных, ролей и аудита", "3.6.3 Контроль данных, ролей и аудита"),
        ("3.5.2 Протокол сценарных проверок", "3.6.2 Протокол сценарных проверок"),
        ("3.5.1 План опытной эксплуатации и приемочные критерии", "3.6.1 План опытной эксплуатации и приемочные критерии"),
        ("3.5 Экономическая и практическая значимость", "3.6 Экономическая и практическая значимость"),
        ("3.4 Квантификация пользовательского интерфейса и метрики проверки", "3.5 Квантификация пользовательского интерфейса и метрики проверки"),
        ("3.3 Анализ результатов разработки", "3.4 Анализ результатов разработки"),
    ]

    for _, paragraph, text in iter_nonempty_paragraphs(doc):
        for old, new in replacements:
            if text == old:
                paragraph.text = new
                break


def update_abstract_counts(doc: Document):
    for _, paragraph, text in iter_nonempty_paragraphs(doc):
        if text.startswith("Пояснительная записка содержит"):
            paragraph.text = re.sub(r"\b33 нумерованные таблицы\b", "51 нумерованную таблицу", paragraph.text)
            paragraph.text = re.sub(r"\b39 рисунков\b", "49 рисунков", paragraph.text)
        if text.startswith("Практическая значимость результата"):
            paragraph.text = paragraph.text.replace("» -", "")


def insert_screen_section(doc: Document):
    target = None
    for _, paragraph, text in iter_nonempty_paragraphs(doc):
        if text == "3.3 Анализ результатов разработки":
            target = paragraph
            break
    if target is None:
        raise RuntimeError("Не найден раздел 3.3 для вставки описания экранов")

    add_before(doc, target, "3.3 Описание основных экранов приложения", "Heading 2")
    add_before(
        doc,
        target,
        "В данном разделе приведено описание ключевых экранов клиентской части информационной системы Пульс. "
        "Раздел оформлен по структуре, использованной в примере Белобородова: для каждого экрана указано его назначение, "
        "приведен экранный снимок и перечислены основные функции, доступные пользователю. Такой формат позволяет связать "
        "проверку работоспособности не только с маршрутом API, но и с фактическим пользовательским интерфейсом.",
    )

    screens = [
        (
            "Экран входа в систему.",
            "01_login.png",
            "Рисунок 34 - Экран входа в систему Пульс",
            [
                "ввод адреса электронной почты и пароля пользователя;",
                "проверка корректности учетных данных на сервере;",
                "переход к рабочей области после успешной авторизации;",
                "отображение сообщения об ошибке при неверных данных.",
            ],
        ),
        (
            "Главная панель пользователя.",
            "02_dashboard.png",
            "Рисунок 35 - Главная панель пользователя",
            [
                "отображение сводной информации по доступным действиям пользователя;",
                "быстрый переход к мероприятиям, задачам, часам и профилю;",
                "учет роли пользователя при формировании доступных разделов.",
            ],
        ),
        (
            "Экран мероприятий.",
            "03_events.png",
            "Рисунок 36 - Экран мероприятий",
            [
                "просмотр списка опубликованных мероприятий организации;",
                "поиск и фильтрация мероприятий по доступным параметрам;",
                "переход к карточке мероприятия и подача заявки на участие.",
            ],
        ),
        (
            "Экран задач.",
            "04_tasks.png",
            "Рисунок 37 - Экран задач волонтера",
            [
                "отображение назначенных задач и их статусов;",
                "просмотр описания задачи и сроков выполнения;",
                "передача результата выполнения в общий контур учета волонтерской деятельности.",
            ],
        ),
        (
            "Экран учета часов.",
            "05_time_entries.png",
            "Рисунок 38 - Экран учета волонтерских часов",
            [
                "создание записи о затраченном времени;",
                "просмотр статуса подтверждения часов координатором;",
                "связь записи времени с мероприятием или задачей.",
            ],
        ),
        (
            "Экран аналитики.",
            "06_analytics.png",
            "Рисунок 39 - Экран аналитики и отчетов",
            [
                "отображение ключевых показателей деятельности организации;",
                "анализ количества мероприятий, задач, активных волонтеров и подтвержденных часов;",
                "подготовка данных для управленческой оценки результатов.",
            ],
        ),
        (
            "Экран сертификатов.",
            "07_certificates.png",
            "Рисунок 40 - Экран сертификатов волонтера",
            [
                "просмотр полученных сертификатов и достижений;",
                "доступ к сведениям о подтвержденном участии;",
                "использование сертификата как результата выполненной волонтерской деятельности.",
            ],
        ),
        (
            "Экран системного администрирования.",
            "08_admin.png",
            "Рисунок 41 - Экран системного администрирования",
            [
                "контроль системных сущностей и справочной информации;",
                "администрирование пользователей и параметров системы;",
                "поддержка разграничения доступа между ролями.",
            ],
        ),
        (
            "Экран настройки полей профиля.",
            "09_profile_fields_admin.png",
            "Рисунок 42 - Экран настройки полей профиля",
            [
                "создание пользовательских групп и полей профиля;",
                "защита системных групп и полей от удаления;",
                "настройка состава данных, собираемых организацией о волонтерах.",
            ],
        ),
        (
            "Экран расширенного профиля волонтера.",
            "10_profile_work_custom_fields.png",
            "Рисунок 43 - Экран расширенного профиля волонтера",
            [
                "просмотр личных данных, задач, работы, мероприятий и достижений;",
                "загрузка аватара пользователя;",
                "заполнение настраиваемых полей профиля в рамках организации.",
            ],
        ),
    ]

    for title, image_name, caption, bullets in screens:
        add_before(doc, target, title)
        add_picture_before(doc, target, SCREENSHOT_DIR / image_name, caption)
        add_before(doc, target, "Функциональность:")
        add_bullets_before(doc, target, bullets)

    add_before(
        doc,
        target,
        "Таким образом, основные экраны приложения подтверждают наличие пользовательского интерфейса для всех ключевых ролей "
        "и операций: входа в систему, просмотра мероприятий, выполнения задач, учета часов, аналитики, сертификатов, "
        "администрирования и настройки профиля.",
    )


def main() -> None:
    doc = Document(INPUT_DOCX)
    insert_screen_section(doc)
    renumber_chapter_three(doc)
    update_abstract_counts(doc)
    doc.save(OUTPUT_DOCX)
    print(OUTPUT_DOCX)


if __name__ == "__main__":
    main()

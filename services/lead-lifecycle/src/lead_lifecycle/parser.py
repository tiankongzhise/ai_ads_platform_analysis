from __future__ import annotations

import csv
import io
import re
import zipfile
from html import unescape
from xml.etree import ElementTree


def parse_tabular(filename: str, content: bytes) -> tuple[list[str], list[dict[str, str]]]:
    lower_name = filename.lower()
    if lower_name.endswith(".xlsx"):
        return parse_xlsx(content)
    return parse_csv(content)


def parse_csv(content: bytes) -> tuple[list[str], list[dict[str, str]]]:
    text = content.decode("utf-8-sig")
    sample = text[:2048]
    dialect = csv.Sniffer().sniff(sample) if sample.strip() else csv.excel
    reader = csv.DictReader(io.StringIO(text), dialect=dialect)
    headers = [normalize_header(item) for item in (reader.fieldnames or [])]
    rows: list[dict[str, str]] = []
    for raw_row in reader:
        row: dict[str, str] = {}
        for original, normalized in zip(reader.fieldnames or [], headers):
            row[normalized] = clean_cell(raw_row.get(original, ""))
        if any(value for value in row.values()):
            rows.append(row)
    return headers, rows


def parse_xlsx(content: bytes) -> tuple[list[str], list[dict[str, str]]]:
    with zipfile.ZipFile(io.BytesIO(content)) as workbook:
        shared_strings = read_shared_strings(workbook)
        sheet_names = workbook.namelist()
        sheet_path = next((name for name in sheet_names if name.startswith("xl/worksheets/sheet") and name.endswith(".xml")), "")
        if not sheet_path:
            return [], []
        root = ElementTree.fromstring(workbook.read(sheet_path))
    rows = []
    for row_el in root.findall(".//{*}sheetData/{*}row"):
        cells: dict[int, str] = {}
        for cell_el in row_el.findall("{*}c"):
            ref = cell_el.attrib.get("r", "")
            col_index = column_index(ref)
            value_el = cell_el.find("{*}v")
            inline_el = cell_el.find("{*}is/{*}t")
            raw_value = value_el.text if value_el is not None else inline_el.text if inline_el is not None else ""
            if cell_el.attrib.get("t") == "s" and raw_value:
                raw_value = shared_strings[int(raw_value)]
            cells[col_index] = clean_cell(raw_value)
        if cells:
            rows.append(cells)
    if not rows:
        return [], []
    header_count = max(rows[0].keys()) + 1
    headers = [normalize_header(rows[0].get(index, "")) for index in range(header_count)]
    records: list[dict[str, str]] = []
    for raw_row in rows[1:]:
        record = {headers[index]: raw_row.get(index, "") for index in range(header_count) if headers[index]}
        if any(value for value in record.values()):
            records.append(record)
    return [header for header in headers if header], records


def read_shared_strings(workbook: zipfile.ZipFile) -> list[str]:
    if "xl/sharedStrings.xml" not in workbook.namelist():
        return []
    root = ElementTree.fromstring(workbook.read("xl/sharedStrings.xml"))
    values = []
    for item in root.findall(".//{*}si"):
        text = "".join(node.text or "" for node in item.findall(".//{*}t"))
        values.append(clean_cell(unescape(text)))
    return values


def normalize_header(value: str | None) -> str:
    return re.sub(r"\s+", "_", clean_cell(value)).strip("_").lower()


def clean_cell(value: object | None) -> str:
    if value is None:
        return ""
    return str(value).strip()


def column_index(cell_ref: str) -> int:
    letters = "".join(ch for ch in cell_ref if ch.isalpha()).upper()
    value = 0
    for ch in letters:
        value = value * 26 + (ord(ch) - ord("A") + 1)
    return max(0, value - 1)


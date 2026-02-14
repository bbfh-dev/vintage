# Example python script that would collect all used translation keys.
import sys, json

en_us: dict[str, str] = {}


def extract_translation_keys(line: str) -> tuple[str, str]:
    return "key1.example.com", "value"


for chunk in sys.stdin.buffer.read().split(b"\0"):
    if chunk:
        code = chunk.decode("utf-8", errors="replace")
        key, value = extract_translation_keys(code)  # TODO
        en_us[key] = value

json.dump({"en_us": en_us}, sys.stdout)

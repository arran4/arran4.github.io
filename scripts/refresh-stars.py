#!/usr/bin/env python3
"""Refresh checked-in public GitHub star counts without credentials.

Transient public API errors must not break deployment or overwrite known counts.
"""
import json
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen

stars_file = Path("data/stars.json")
cached = json.loads(stars_file.read_text(encoding="utf-8"))
updated = dict(cached)

for repository in sorted(cached):
    if not repository.startswith("arran4/"):
        print(f"::warning::Skipping unexpected repository {repository!r}")
        continue
    url = "https://api.github.com/repos/" + quote(repository, safe="/")
    try:
        request = Request(url, headers={"User-Agent": "arran4-github-pages-public-metadata"})
        with urlopen(request, timeout=12) as response:
            count = json.load(response)["stargazers_count"]
        if not isinstance(count, int) or isinstance(count, bool) or count < 0:
            raise ValueError("Invalid public stargazers_count")
        updated[repository] = count
    except (HTTPError, URLError, TimeoutError, ValueError, KeyError) as exc:
        print(f"::warning::Public star count unavailable for {repository}: {exc}. Using cached value.")

if updated != cached:
    stars_file.write_text(json.dumps(updated, indent=2) + "\n", encoding="utf-8")

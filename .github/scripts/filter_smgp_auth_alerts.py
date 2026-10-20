"""Apply the three reviewed SMGP wire-authentication exceptions to CodeQL SARIF.

SMGP 3.0.3 sections 6.2.2 and 6.2.7 require MD5. These examples compute
protocol authenticators, not stored password hashes; replacing MD5 would break
interoperability. The legacy protocol still needs a trusted/encrypted transport.
Keep every other finding, including other weak hashes in these same files.
"""

import argparse
import json
from pathlib import Path, PurePosixPath
from urllib.parse import unquote, urlsplit


RULE = "go/weak-sensitive-data-hashing"
EXCEPTIONS = {
    "nioserver/examples/smgp30server/auth.go": {
        "return md5.Sum(data)": "// SMGP protocol exception: 3.0.3 section 6.2.2 requires MD5 for AuthenticatorClient.",
        "auth := md5.Sum(data)": "// SMGP protocol exception: 3.0.3 section 6.2.7 requires MD5 for AuthenticatorServer.",
    },
    "nioserver/examples/smgp30client/main.go": {
        "want := md5.Sum(data)": "// SMGP protocol exception: 3.0.3 section 6.2.7 requires MD5 to verify AuthenticatorServer.",
    },
}


def protocol_exception(result, source_root):
    if result.get("ruleId") != RULE or len(result.get("locations", [])) != 1:
        return None
    location = result["locations"][0].get("physicalLocation", {})
    uri = urlsplit(location.get("artifactLocation", {}).get("uri", ""))
    if uri.scheme or uri.netloc or uri.query or uri.fragment:
        return None
    path = unquote(uri.path)
    expressions = EXCEPTIONS.get(path)
    if expressions is None or ".." in PurePosixPath(path).parts:
        return None
    region = location.get("region", {})
    line = region.get("startLine", 0)
    if not isinstance(line, int) or line < 2 or region.get("endLine", line) != line:
        return None
    source = (source_root / path).resolve()
    if not source.is_relative_to(source_root.resolve()) or not source.is_file():
        return None
    lines = source.read_text(encoding="utf-8").splitlines()
    if line > len(lines):
        return None
    annotation = expressions.get(lines[line - 1].strip())
    if annotation is None or lines[line - 2].strip() != annotation:
        return None
    return f"{path}:{line}: {annotation.removeprefix('// ')}"


def filter_results(document, source_root):
    accepted = []
    for run in document["runs"]:
        if run.get("tool", {}).get("driver", {}).get("name") != "CodeQL":
            continue
        kept = []
        for result in run.get("results", []):
            exception = protocol_exception(result, source_root)
            if exception is None:
                kept.append(result)
            else:
                accepted.append(exception)
        run["results"] = kept
    return accepted


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("input", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--source-root", type=Path, default=Path.cwd())
    args = parser.parse_args()
    document = json.loads(args.input.read_text(encoding="utf-8"))
    for exception in filter_results(document, args.source_root):
        print(f"Reviewed protocol exception ({RULE}): {exception}")
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(document, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()

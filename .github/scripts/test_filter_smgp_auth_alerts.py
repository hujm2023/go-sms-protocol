import copy
from pathlib import Path
import tempfile
import unittest

from filter_smgp_auth_alerts import EXCEPTIONS, RULE, filter_results


def alert(path, line, rule=RULE):
    return {
        "ruleId": rule,
        "message": {"text": "finding"},
        "partialFingerprints": {"primaryLocationLineHash": "preserved"},
        "locations": [{"physicalLocation": {
            "artifactLocation": {"uri": path},
            "region": {"startLine": line, "endLine": line},
        }}],
    }


def sarif(results):
    return {"version": "2.1.0", "runs": [{
        "tool": {"driver": {"name": "CodeQL"}},
        "automationDetails": {"id": "/job:codeql/language:go/"},
        "results": results,
    }]}


class ProtocolExceptionsTest(unittest.TestCase):
    def setUp(self):
        self.root = Path(__file__).resolve().parents[2]

    def reviewed_alerts(self):
        findings = []
        for path, expressions in EXCEPTIONS.items():
            lines = (self.root / path).read_text(encoding="utf-8").splitlines()
            for expression, annotation in expressions.items():
                candidates = [n + 1 for n, line in enumerate(lines)
                              if line.strip() == expression and n > 0
                              and lines[n - 1].strip() == annotation]
                self.assertEqual(len(candidates), 1, (path, expression))
                findings.append(alert(path, candidates[0]))
        return findings

    def test_only_three_documented_calls_are_excepted(self):
        reviewed = self.reviewed_alerts()
        self.assertEqual(len(reviewed), 3)
        other_rule = alert(reviewed[0]["locations"][0]["physicalLocation"]["artifactLocation"]["uri"],
                           reviewed[0]["locations"][0]["physicalLocation"]["region"]["startLine"],
                           "go/clear-text-logging")
        unreviewed = alert("nioserver/examples/smgp30server/auth.go", 1)
        elsewhere = alert("smgp/smgp30/utils.go", 109)
        untouched = [other_rule, unreviewed, elsewhere]
        document = sarif(reviewed + copy.deepcopy(untouched))
        before = copy.deepcopy(document["runs"][0]["automationDetails"])
        self.assertEqual(len(filter_results(document, self.root)), 3)
        self.assertEqual(document["runs"][0]["results"], untouched)
        self.assertEqual(document["runs"][0]["automationDetails"], before)

    def test_missing_annotation_or_changed_expression_keeps_finding(self):
        path = "nioserver/examples/smgp30server/auth.go"
        annotation = EXCEPTIONS[path]["return md5.Sum(data)"]
        for name, source in [
            ("no annotation", "// unrelated\nreturn md5.Sum(data)\n"),
            ("changed expression", annotation + "\nreturn md5.Sum(password)\n"),
        ]:
            with self.subTest(name=name), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                file = root / path
                file.parent.mkdir(parents=True)
                file.write_text(source, encoding="utf-8")
                document = sarif([alert(path, 2)])
                before = copy.deepcopy(document)
                self.assertEqual(filter_results(document, root), [])
                self.assertEqual(document, before)

    def test_ambiguous_locations_and_other_tools_are_preserved(self):
        finding = self.reviewed_alerts()[0]
        finding["locations"].append(copy.deepcopy(finding["locations"][0]))
        document = sarif([finding])
        before = copy.deepcopy(document)
        self.assertEqual(filter_results(document, self.root), [])
        self.assertEqual(document, before)
        document = sarif(self.reviewed_alerts())
        document["runs"][0]["tool"]["driver"]["name"] = "Other scanner"
        before = copy.deepcopy(document)
        self.assertEqual(filter_results(document, self.root), [])
        self.assertEqual(document, before)


if __name__ == "__main__":
    unittest.main()

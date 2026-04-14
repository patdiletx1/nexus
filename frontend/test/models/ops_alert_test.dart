import "package:flutter_test/flutter_test.dart";

import "package:nexus_frontend/models/ops_alert.dart";

void main() {
  group("OpsAlert.fromFormattedResponse", () {
    test("returns empty list for non-200 response", () {
      final List<OpsAlert> alerts = OpsAlert.fromFormattedResponse(
        "HTTP 401\n{\"error\":\"unauthorized\"}",
      );
      expect(alerts, isEmpty);
    });

    test("returns empty list for invalid json payload", () {
      final List<OpsAlert> alerts = OpsAlert.fromFormattedResponse(
        "HTTP 200\nnot-json",
      );
      expect(alerts, isEmpty);
    });

    test("sorts by triggered first then severity", () {
      const String formatted = """
HTTP 200
{
  "alerts": [
    {
      "name": "warn_not_triggered",
      "severity": "warning",
      "triggered": false,
      "description": "warning"
    },
    {
      "name": "critical_triggered",
      "severity": "critical",
      "triggered": true,
      "description": "critical"
    },
    {
      "name": "warning_triggered",
      "severity": "warning",
      "triggered": true,
      "description": "warning"
    },
    {
      "name": "critical_not_triggered",
      "severity": "critical",
      "triggered": false,
      "description": "critical"
    }
  ]
}
""";

      final List<OpsAlert> alerts = OpsAlert.fromFormattedResponse(formatted);
      expect(alerts.map((a) => a.name).toList(), <String>[
        "critical_triggered",
        "warning_triggered",
        "critical_not_triggered",
        "warn_not_triggered",
      ]);
    });
  });
}

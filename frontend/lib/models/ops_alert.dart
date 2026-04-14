import "dart:convert";

class OpsAlert {
  const OpsAlert({
    required this.name,
    required this.severity,
    required this.triggered,
    required this.description,
  });

  final String name;
  final String severity;
  final bool triggered;
  final String description;

  int get severityRank => switch (severity.toLowerCase()) {
        "critical" => 0,
        "warning" => 1,
        _ => 2,
      };

  static List<OpsAlert> fromFormattedResponse(String formattedResponse) {
    final String raw = formattedResponse.trim();
    if (raw.isEmpty) {
      return const <OpsAlert>[];
    }
    final List<String> lines = raw.split("\n");
    if (lines.isEmpty || !lines.first.startsWith("HTTP 200")) {
      return const <OpsAlert>[];
    }
    final String jsonRaw = lines.skip(1).join("\n").trim();
    if (jsonRaw.isEmpty) {
      return const <OpsAlert>[];
    }
    try {
      final Map<String, dynamic> payload = jsonDecode(jsonRaw) as Map<String, dynamic>;
      final List<dynamic> alerts = payload["alerts"] as List<dynamic>? ?? const <dynamic>[];
      final List<OpsAlert> mapped = alerts
          .whereType<Map<String, dynamic>>()
          .map(
            (Map<String, dynamic> alert) => OpsAlert(
              name: (alert["name"] ?? "").toString(),
              severity: (alert["severity"] ?? "").toString(),
              triggered: alert["triggered"] == true,
              description: (alert["description"] ?? "").toString(),
            ),
          )
          .toList();

      mapped.sort((a, b) {
        if (a.triggered != b.triggered) {
          return a.triggered ? -1 : 1;
        }
        final int severityCompare = a.severityRank.compareTo(b.severityRank);
        if (severityCompare != 0) {
          return severityCompare;
        }
        return a.name.compareTo(b.name);
      });
      return mapped;
    } catch (_) {
      return const <OpsAlert>[];
    }
  }
}

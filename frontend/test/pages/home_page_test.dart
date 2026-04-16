import "package:flutter/material.dart";
import "package:flutter_test/flutter_test.dart";
import "package:shared_preferences/shared_preferences.dart";

import "package:nexus_frontend/pages/home_page.dart";
import "package:nexus_frontend/services/nexus_api_client.dart";

class _FakeNexusApiClient extends NexusApiClient {
  const _FakeNexusApiClient();

  static bool called = false;

  @override
  Future<String> getOpsAlerts({
    required String baseUrl,
    required String token,
  }) async {
    called = true;
    return """HTTP 200
{
  "alerts": [
    {
      "name": "warning_not_triggered",
      "severity": "warning",
      "triggered": false,
      "description": "warn"
    },
    {
      "name": "critical_triggered",
      "severity": "critical",
      "triggered": true,
      "description": "crit"
    }
  ]
}
""";
  }
}

void main() {
  testWidgets("renders sorted ops alert chips from response", (tester) async {
    _FakeNexusApiClient.called = false;
    SharedPreferences.setMockInitialValues(<String, Object>{
      "nexus.base_url": "http://localhost:8080",
      "nexus.jwt_token": "local-token",
    });

    await tester.pumpWidget(
      const MaterialApp(
        home: HomePage(client: _FakeNexusApiClient()),
      ),
    );
    await tester.pumpAndSettle();

    final Finder opsButton = find.widgetWithText(FilledButton, "Ops Alerts");
    expect(opsButton, findsOneWidget);
    await tester.ensureVisible(opsButton);
    await tester.tap(opsButton);
    await tester.pumpAndSettle();

    expect(_FakeNexusApiClient.called, isTrue);
    expect(find.text("JWT token requerido para esta accion"), findsNothing);

    final Finder overviewOffstageFinder = find.text(
      "Ops alerts overview",
      skipOffstage: false,
    );
    expect(overviewOffstageFinder, findsOneWidget);
    await tester.scrollUntilVisible(
      overviewOffstageFinder,
      300,
      scrollable: find.byType(Scrollable).first,
    );
    await tester.pumpAndSettle();

    expect(find.text("Ops alerts overview"), findsOneWidget);
    expect(find.text("critical_triggered"), findsOneWidget);
    expect(find.text("warning_not_triggered"), findsOneWidget);
    expect(find.text("critical"), findsOneWidget);
    expect(find.text("warning"), findsOneWidget);
    expect(find.text("triggered"), findsOneWidget);
    expect(find.text("ok"), findsOneWidget);

    final double criticalY = tester.getTopLeft(find.text("critical_triggered")).dy;
    final double warningY = tester.getTopLeft(find.text("warning_not_triggered")).dy;
    expect(criticalY, lessThan(warningY));
  });
}

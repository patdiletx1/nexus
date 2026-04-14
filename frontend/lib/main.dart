import "dart:convert";

import "package:flutter/material.dart";
import "package:http/http.dart" as http;

void main() {
  runApp(const NexusApp());
}

class NexusApp extends StatelessWidget {
  const NexusApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: "Nexus Local",
      theme: ThemeData(useMaterial3: true, colorSchemeSeed: Colors.indigo),
      home: const NexusHomePage(),
    );
  }
}

class NexusHomePage extends StatefulWidget {
  const NexusHomePage({super.key});

  @override
  State<NexusHomePage> createState() => _NexusHomePageState();
}

class _NexusHomePageState extends State<NexusHomePage> {
  final TextEditingController _baseUrlController = TextEditingController(
    text: "http://localhost:8080",
  );
  final TextEditingController _tokenController = TextEditingController();

  String _healthResponse = "";
  String _profileResponse = "";
  String _tendersResponse = "";
  String _errorMessage = "";
  bool _loading = false;

  @override
  void dispose() {
    _baseUrlController.dispose();
    _tokenController.dispose();
    super.dispose();
  }

  Future<void> _runRequest({
    required String path,
    required bool requiresAuth,
    required ValueSetter<String> onSuccess,
  }) async {
    setState(() {
      _loading = true;
      _errorMessage = "";
    });

    try {
      final String baseUrl = _baseUrlController.text.trim().replaceAll(RegExp(r"/+$"), "");
      final Uri uri = Uri.parse("$baseUrl$path");
      final Map<String, String> headers = {"Accept": "application/json"};

      if (requiresAuth) {
        final String token = _tokenController.text.trim();
        if (token.isEmpty) {
          throw Exception("JWT token requerido para este endpoint");
        }
        headers["Authorization"] = "Bearer $token";
      }

      final http.Response resp = await http.get(uri, headers: headers);
      final String prettyBody = _prettyJsonOrRaw(resp.body);
      onSuccess("HTTP ${resp.statusCode}\n$prettyBody");
    } catch (err) {
      setState(() {
        _errorMessage = err.toString();
      });
    } finally {
      setState(() {
        _loading = false;
      });
    }
  }

  String _prettyJsonOrRaw(String raw) {
    try {
      final dynamic decoded = jsonDecode(raw);
      return const JsonEncoder.withIndent("  ").convert(decoded);
    } catch (_) {
      return raw;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("Nexus Frontend Bootstrap")),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: ListView(
          children: <Widget>[
            TextField(
              controller: _baseUrlController,
              decoration: const InputDecoration(
                labelText: "API_BASE_URL",
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: _tokenController,
              minLines: 2,
              maxLines: 3,
              decoration: const InputDecoration(
                labelText: "JWT_TOKEN (Bearer)",
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: <Widget>[
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () => _runRequest(
                            path: "/health/live",
                            requiresAuth: false,
                            onSuccess: (value) => setState(() => _healthResponse = value),
                          ),
                  child: const Text("Health"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () => _runRequest(
                            path: "/v1/company/profile",
                            requiresAuth: true,
                            onSuccess: (value) => setState(() => _profileResponse = value),
                          ),
                  child: const Text("Profile"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () => _runRequest(
                            path: "/v1/tenders?limit=20",
                            requiresAuth: true,
                            onSuccess: (value) => setState(() => _tendersResponse = value),
                          ),
                  child: const Text("Tenders"),
                ),
              ],
            ),
            if (_errorMessage.isNotEmpty) ...<Widget>[
              const SizedBox(height: 12),
              Text(
                _errorMessage,
                style: const TextStyle(color: Colors.red),
              ),
            ],
            const SizedBox(height: 16),
            _ResponseCard(title: "Health response", content: _healthResponse),
            const SizedBox(height: 12),
            _ResponseCard(title: "Profile response", content: _profileResponse),
            const SizedBox(height: 12),
            _ResponseCard(title: "Tenders response", content: _tendersResponse),
          ],
        ),
      ),
    );
  }
}

class _ResponseCard extends StatelessWidget {
  const _ResponseCard({required this.title, required this.content});

  final String title;
  final String content;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(title, style: Theme.of(context).textTheme.titleSmall),
            const SizedBox(height: 8),
            SelectableText(content.isEmpty ? "Sin datos aun." : content),
          ],
        ),
      ),
    );
  }
}

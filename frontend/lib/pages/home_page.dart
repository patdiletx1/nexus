import "package:flutter/material.dart";

import "../services/nexus_api_client.dart";
import "../widgets/response_card.dart";

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final NexusApiClient _client = const NexusApiClient();
  final TextEditingController _baseUrlController = TextEditingController(
    text: "http://localhost:8080",
  );
  final TextEditingController _tokenController = TextEditingController();
  final TextEditingController _tenderIdController = TextEditingController(
    text: "MOCK-003",
  );

  String _healthResponse = "";
  String _profileResponse = "";
  String _syncResponse = "";
  String _tendersResponse = "";
  String _warmupResponse = "";
  String _scoreResponse = "";
  String _errorMessage = "";
  bool _loading = false;

  @override
  void dispose() {
    _baseUrlController.dispose();
    _tokenController.dispose();
    _tenderIdController.dispose();
    super.dispose();
  }

  Future<void> _run({
    required Future<String> Function() action,
    required ValueSetter<String> onSuccess,
  }) async {
    setState(() {
      _loading = true;
      _errorMessage = "";
    });
    try {
      final String result = await action();
      setState(() {
        onSuccess(result);
      });
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

  String get _baseUrl => _baseUrlController.text.trim();
  String get _token => _tokenController.text.trim();

  bool _requireToken() {
    if (_token.isNotEmpty) {
      return true;
    }
    setState(() {
      _errorMessage = "JWT token requerido para esta accion";
    });
    return false;
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
            TextField(
              controller: _tenderIdController,
              decoration: const InputDecoration(
                labelText: "Tender ID para score",
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
                      : () => _run(
                            action: () => _client.getHealth(_baseUrl),
                            onSuccess: (value) => _healthResponse = value,
                          ),
                  child: const Text("Health"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () {
                          if (!_requireToken()) return;
                          _run(
                            action: () => _client.getProfile(
                              baseUrl: _baseUrl,
                              token: _token,
                            ),
                            onSuccess: (value) => _profileResponse = value,
                          );
                        },
                  child: const Text("Profile"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () {
                          if (!_requireToken()) return;
                          _run(
                            action: () => _client.syncTenders(
                              baseUrl: _baseUrl,
                              token: _token,
                            ),
                            onSuccess: (value) => _syncResponse = value,
                          );
                        },
                  child: const Text("Sync"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () {
                          if (!_requireToken()) return;
                          _run(
                            action: () => _client.listTenders(
                              baseUrl: _baseUrl,
                              token: _token,
                            ),
                            onSuccess: (value) => _tendersResponse = value,
                          );
                        },
                  child: const Text("Tenders"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () {
                          if (!_requireToken()) return;
                          _run(
                            action: () => _client.warmup(
                              baseUrl: _baseUrl,
                              token: _token,
                            ),
                            onSuccess: (value) => _warmupResponse = value,
                          );
                        },
                  child: const Text("Warmup"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () {
                          if (!_requireToken()) return;
                          final String tenderId = _tenderIdController.text.trim();
                          if (tenderId.isEmpty) {
                            setState(() {
                              _errorMessage = "Tender ID requerido para score";
                            });
                            return;
                          }
                          _run(
                            action: () => _client.scoreTender(
                              baseUrl: _baseUrl,
                              token: _token,
                              tenderId: tenderId,
                            ),
                            onSuccess: (value) => _scoreResponse = value,
                          );
                        },
                  child: const Text("Score"),
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
            ResponseCard(title: "Health response", content: _healthResponse),
            const SizedBox(height: 12),
            ResponseCard(title: "Profile response", content: _profileResponse),
            const SizedBox(height: 12),
            ResponseCard(title: "Sync response", content: _syncResponse),
            const SizedBox(height: 12),
            ResponseCard(title: "Tenders response", content: _tendersResponse),
            const SizedBox(height: 12),
            ResponseCard(title: "Warmup response", content: _warmupResponse),
            const SizedBox(height: 12),
            ResponseCard(title: "Score response", content: _scoreResponse),
          ],
        ),
      ),
    );
  }
}

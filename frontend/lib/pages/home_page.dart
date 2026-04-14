import "package:flutter/material.dart";
import "package:shared_preferences/shared_preferences.dart";

import "../services/nexus_api_client.dart";
import "../widgets/response_card.dart";

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  static const String _baseUrlKey = "nexus.base_url";
  static const String _jwtTokenKey = "nexus.jwt_token";
  static const String _tenderIdKey = "nexus.tender_id";

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
  String _opsAlertsResponse = "";
  String _metricsResponse = "";
  String _errorMessage = "";
  bool _loading = false;
  bool _loadingSavedInputs = true;
  bool _obscureToken = true;

  @override
  void initState() {
    super.initState();
    _baseUrlController.addListener(_persistInputs);
    _tokenController.addListener(_persistInputs);
    _tenderIdController.addListener(_persistInputs);
    _loadPersistedInputs();
  }

  Future<void> _loadPersistedInputs() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String? savedBaseUrl = prefs.getString(_baseUrlKey);
    final String? savedToken = prefs.getString(_jwtTokenKey);
    final String? savedTenderId = prefs.getString(_tenderIdKey);
    if (!mounted) return;
    setState(() {
      if ((savedBaseUrl ?? "").trim().isNotEmpty) {
        _baseUrlController.text = savedBaseUrl!.trim();
      }
      if ((savedToken ?? "").isNotEmpty) {
        _tokenController.text = savedToken!;
      }
      if ((savedTenderId ?? "").trim().isNotEmpty) {
        _tenderIdController.text = savedTenderId!.trim();
      }
      _loadingSavedInputs = false;
    });
  }

  Future<void> _persistInputs() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    await prefs.setString(_baseUrlKey, _baseUrlController.text.trim());
    await prefs.setString(_jwtTokenKey, _tokenController.text);
    await prefs.setString(_tenderIdKey, _tenderIdController.text.trim());
  }

  Future<void> _clearLocalSession() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    await prefs.remove(_baseUrlKey);
    await prefs.remove(_jwtTokenKey);
    await prefs.remove(_tenderIdKey);
    if (!mounted) return;
    setState(() {
      _baseUrlController.text = "http://localhost:8080";
      _tokenController.clear();
      _tenderIdController.text = "MOCK-003";
      _healthResponse = "";
      _profileResponse = "";
      _syncResponse = "";
      _tendersResponse = "";
      _warmupResponse = "";
      _scoreResponse = "";
      _opsAlertsResponse = "";
      _metricsResponse = "";
      _errorMessage = "";
    });
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text("Sesion local limpiada")),
    );
  }

  Future<void> _confirmAndClearLocalSession() async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: const Text("Limpiar sesion local"),
          content: const Text(
            "Se borraran API_BASE_URL, JWT_TOKEN y Tender ID guardados localmente.",
          ),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(false),
              child: const Text("Cancelar"),
            ),
            FilledButton(
              onPressed: () => Navigator.of(dialogContext).pop(true),
              child: const Text("Limpiar"),
            ),
          ],
        );
      },
    );
    if (confirmed != true) {
      return;
    }
    await _clearLocalSession();
  }

  @override
  void dispose() {
    _baseUrlController.removeListener(_persistInputs);
    _tokenController.removeListener(_persistInputs);
    _tenderIdController.removeListener(_persistInputs);
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
            if (_loadingSavedInputs)
              const Padding(
                padding: EdgeInsets.only(bottom: 12),
                child: LinearProgressIndicator(),
              ),
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
              obscureText: _obscureToken,
              decoration: InputDecoration(
                labelText: "JWT_TOKEN (Bearer)",
                border: const OutlineInputBorder(),
                suffixIcon: IconButton(
                  tooltip: _obscureToken ? "Mostrar token" : "Ocultar token",
                  onPressed: () {
                    setState(() {
                      _obscureToken = !_obscureToken;
                    });
                  },
                  icon: Icon(
                    _obscureToken ? Icons.visibility : Icons.visibility_off,
                  ),
                ),
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
            OutlinedButton.icon(
              onPressed: _loading ? null : _confirmAndClearLocalSession,
              icon: const Icon(Icons.delete_sweep_outlined),
              label: const Text("Limpiar sesion local"),
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
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () {
                          if (!_requireToken()) return;
                          _run(
                            action: () => _client.getOpsAlerts(
                              baseUrl: _baseUrl,
                              token: _token,
                            ),
                            onSuccess: (value) => _opsAlertsResponse = value,
                          );
                        },
                  child: const Text("Ops Alerts"),
                ),
                FilledButton(
                  onPressed: _loading
                      ? null
                      : () => _run(
                            action: () => _client.getMetricsSummary(
                              baseUrl: _baseUrl,
                            ),
                            onSuccess: (value) => _metricsResponse = value,
                          ),
                  child: const Text("Metrics"),
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
            const SizedBox(height: 12),
            ResponseCard(title: "Ops alerts response", content: _opsAlertsResponse),
            const SizedBox(height: 12),
            ResponseCard(title: "Metrics summary", content: _metricsResponse),
          ],
        ),
      ),
    );
  }
}

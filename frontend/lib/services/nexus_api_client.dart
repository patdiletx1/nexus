import "dart:convert";

import "package:http/http.dart" as http;

class NexusApiClient {
  const NexusApiClient();

  Future<String> getHealth(String baseUrl) {
    return _get(baseUrl: baseUrl, path: "/health/live", token: null);
  }

  Future<String> getProfile({
    required String baseUrl,
    required String token,
  }) {
    return _get(
      baseUrl: baseUrl,
      path: "/v1/company/profile",
      token: token,
    );
  }

  Future<String> syncTenders({
    required String baseUrl,
    required String token,
    int limit = 20,
  }) {
    return _get(
      baseUrl: baseUrl,
      path: "/v1/tenders/sync?limit=$limit",
      token: token,
    );
  }

  Future<String> listTenders({
    required String baseUrl,
    required String token,
    int limit = 20,
  }) {
    return _get(
      baseUrl: baseUrl,
      path: "/v1/tenders?limit=$limit",
      token: token,
    );
  }

  Future<String> warmup({
    required String baseUrl,
    required String token,
    int limit = 20,
  }) async {
    return _post(
      baseUrl: baseUrl,
      path: "/v1/tenders/score/warmup",
      token: token,
      body: {"limit": limit},
    );
  }

  Future<String> scoreTender({
    required String baseUrl,
    required String token,
    required String tenderId,
  }) {
    return _get(
      baseUrl: baseUrl,
      path: "/v1/tenders/$tenderId/score",
      token: token,
    );
  }

  Future<String> _get({
    required String baseUrl,
    required String path,
    required String? token,
  }) async {
    final Uri uri = Uri.parse(_normalizedBase(baseUrl) + path);
    final Map<String, String> headers = _headers(token);
    final http.Response response = await http.get(uri, headers: headers);
    return _formatted(response);
  }

  Future<String> _post({
    required String baseUrl,
    required String path,
    required String token,
    required Map<String, dynamic> body,
  }) async {
    final Uri uri = Uri.parse(_normalizedBase(baseUrl) + path);
    final Map<String, String> headers = _headers(token)
      ..putIfAbsent("Content-Type", () => "application/json");
    final http.Response response = await http.post(
      uri,
      headers: headers,
      body: jsonEncode(body),
    );
    return _formatted(response);
  }

  String _normalizedBase(String baseUrl) {
    return baseUrl.trim().replaceAll(RegExp(r"/+$"), "");
  }

  Map<String, String> _headers(String? token) {
    final Map<String, String> headers = {"Accept": "application/json"};
    final String safeToken = (token ?? "").trim();
    if (safeToken.isNotEmpty) {
      headers["Authorization"] = "Bearer $safeToken";
    }
    return headers;
  }

  String _formatted(http.Response response) {
    return "HTTP ${response.statusCode}\n${_prettyJsonOrRaw(response.body)}";
  }

  String _prettyJsonOrRaw(String raw) {
    try {
      final dynamic decoded = jsonDecode(raw);
      return const JsonEncoder.withIndent("  ").convert(decoded);
    } catch (_) {
      return raw;
    }
  }
}

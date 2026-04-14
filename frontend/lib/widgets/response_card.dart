import "package:flutter/material.dart";

class ResponseCard extends StatelessWidget {
  const ResponseCard({
    required this.title,
    required this.content,
    super.key,
  });

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

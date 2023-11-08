class Configuration {
  final String name;

  const Configuration({required this.name});

  factory Configuration.fromJson(Map<String, dynamic> json) {
    return Configuration(name: json['name'] as String);
  }
}

class RenderResponse {
  final Map<String, dynamic> viewModel;
  final VView renderTree;

  const RenderResponse({required this.viewModel, required this.renderTree});

  factory RenderResponse.fromJson(Map<String, dynamic> json) {
    final view = json["renderTree"] as dynamic;
    return RenderResponse(
        viewModel: json['viewModel'], renderTree: VView.fromJson(view));
  }
}

class VView {
  final String type;

  const VView({required this.type});

  factory VView.fromJson(Map<String, dynamic> json) {
    switch (json["type"] as String) {
      case "Text":
        return VText.fromJson(json);
      case "TableCell":
        return VTableCell.fromJson(json);
      case "TableRow":
        return VTableRow.fromJson(json);
      case "Table":
        return VTable.fromJson(json);
      case "TableColumnHeader":
        return VTableColumnHeader.fromJson(json);
      default:
        return VText(
            type: "Text", value: "unsupported model REST type $json['type']");
    }
  }

  static List<VView> fromJsonArray(List<dynamic> json) {
    return json.map((e) => VView.fromJson(e as Map<String, dynamic>)).toList();
  }
}

class VText extends VView {
  final String value;

  const VText({required super.type, required this.value});

  factory VText.fromJson(Map<String, dynamic> json) {
    return VText(type: json['type'] as String, value: json['value'] as String);
  }
}

class VTable extends VView {
  final List<VTableColumnHeader> headers;
  final List<VTableRow> rows;

  const VTable(
      {required super.type, required this.headers, required this.rows});

  factory VTable.fromJson(Map<String, dynamic> json) {
    final views = VView.fromJsonArray(json["rows"] as List<dynamic>);
    final List<VTableRow> rows = List.from(views);

    final headers = VView.fromJsonArray(json["columnHeaders"] as List<dynamic>);
    final List<VTableColumnHeader> vheaders = List.from(headers);

    return VTable(type: json['type'] as String, rows: rows, headers: vheaders);
  }
}

class VTableRow extends VView {
  final List<VTableCell> columns;

  const VTableRow({required super.type, required this.columns});

  factory VTableRow.fromJson(Map<String, dynamic> json) {
    final views = VView.fromJsonArray(json["columns"] as List<dynamic>);
    final List<VTableCell> cells = List.from(views);
    return VTableRow(type: json['type'] as String, columns: cells);
  }
}

class VTableCell extends VView {
  final List<VView> views;

  const VTableCell({required super.type, required this.views});

  factory VTableCell.fromJson(Map<String, dynamic> json) {
    return VTableCell(
        type: json['type'] as String,
        views: VView.fromJsonArray(json["views"] as List<dynamic>));
  }
}

class VTableColumnHeader extends VView {
  final List<VView> views;

  const VTableColumnHeader({required super.type, required this.views});

  factory VTableColumnHeader.fromJson(Map<String, dynamic> json) {
    return VTableColumnHeader(
        type: json['type'] as String,
        views: VView.fromJsonArray(json["views"] as List<dynamic>));
  }
}

class VScaffold extends VView {
  final String title;
  final List<LabelIconItem> menuItems;
  final Body VView;
}

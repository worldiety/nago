import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:nago_frontend/model.dart';
import 'package:flutter_adaptive_scaffold/flutter_adaptive_scaffold.dart';

class Render {
  static MaterialApp makeApp(Widget w) {
    return MaterialApp(
      title: "invalid app state",
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.red),
        useMaterial3: true,
      ),
      home: w,
    );
  }

  static AdaptiveScaffold renderScaffold(VScaffhold scaffhold) {}

  static DataTable renderTable(VTable table) {
    return DataTable(
        columns: renderTableHeader(table.headers),
        rows: renderTableRows(table.rows));
  }

  static Widget renderView(VView? v) {
    if (v == null) {
      return const Text("render tree is null");
    }

    switch (v.runtimeType) {
      case VText:
        return Text((v as VText).value);
      case VTable:
        final table = v as VTable;
        return DataTable(
            columns: renderTableHeader(table.headers),
            rows: renderTableRows(table.rows));
      default:
        final rt = v.runtimeType;
        final t = v.type;
        return Text("rendering of class '$rt' not implemented (type=$t)");
    }
  }

  static List<DataColumn> renderTableHeader(List<VTableColumnHeader> headers) {
    return headers
        .map((e) => DataColumn(label: renderView(e.views.first)))
        .toList();
  }

  static List<DataRow> renderTableRows(List<VTableRow> rows) {
    return rows
        .map((e) => DataRow(
            cells: e.columns
                .map((e) => DataCell(renderView(e.views.first)))
                .toList()))
        .toList();
  }
}

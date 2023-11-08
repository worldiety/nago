import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

import 'package:nago_frontend/model.dart';
import 'package:nago_frontend/render.dart';

void main() {
  runApp(const DynamicPage());
}

Future<Configuration> fetchConfiguration() async {
  final resp = await http
      .get(Uri.parse("http://localhost:3000/api/v1/ui/configuration"));
  return Configuration.fromJson(jsonDecode(resp.body) as Map<String, dynamic>);
}

class DynamicPage extends StatefulWidget {
  const DynamicPage();

  @override
  State<DynamicPage> createState() => _DynamicPageState();
}

class _DynamicPageState extends State<DynamicPage> {
  late Future<Configuration> futureConfig;

  VView? _view;

  void updateRenderTree(VView view) {
    setState(() {
      _view = view;
    });
  }

  @override
  void initState() {
    super.initState();
    futureConfig = fetchConfiguration();
    showPage("events/public");
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder(
        future: futureConfig,
        builder: (context, snapshot) {
          if (snapshot.hasData) {
            final bool displayMobileLayout =
                MediaQuery.of(context).size.width < 800;

            final cfg = snapshot.data!;
            final app = MaterialApp(
              title: cfg.name,
              theme: ThemeData(
                colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
                useMaterial3: true,
              ),
              home: Scaffold(
                appBar: AppBar(
                  title: Text("yo"),
                  backgroundColor: Theme.of(context).colorScheme.inversePrimary,
                ),
                body: () {
                  if (!displayMobileLayout) {
                    return Row(
                      children: [
                        Drawer(
                            child: ListView(
                          children: [
                            ListTile(
                                leading: Icon(Icons.message),
                                title: Text("bla"))
                          ],
                        )),
                        Render.renderView(_view)
                      ],
                    );
                  } else {
                    return Render.renderView(_view);
                  }
                }(),
                drawer: Drawer(
                    child: ListView(
                  children: [
                    ListTile(
                      leading: Icon(Icons.message),
                      title: Text('Messages a'),
                    )
                  ],
                )),
              ),
            );

            return app;
          } else if (snapshot.hasError) {
            print("error: $snapshot.error");
            return Render.makeApp(Text("ERROR: $snapshot.error"));
          }

          print("no data or error");

          return const CircularProgressIndicator();
        });
  }

  void showPage(String pageID) {
    print("should show $pageID");
    http
        .get(Uri.parse("http://localhost:3000/api/v1/ui/page/$pageID"))
        .then((res) {
      final rr =
          RenderResponse.fromJson(jsonDecode(res.body) as Map<String, dynamic>);
      updateRenderTree(rr.renderTree);
    });
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  // This widget is the home page of your application. It is stateful, meaning
  // that it has a State object (defined below) that contains fields that affect
  // how it looks.

  // This class is the configuration for the state. It holds the values (in this
  // case the title) provided by the parent (in this case the App widget) and
  // used by the build method of the State. Fields in a Widget subclass are
  // always marked "final".

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      // This call to setState tells the Flutter framework that something has
      // changed in this State, which causes it to rerun the build method below
      // so that the display can reflect the updated values. If we changed
      // _counter without calling setState(), then the build method would not be
      // called again, and so nothing would appear to happen.
      _counter++;
    });
  }

  @override
  Widget build(BuildContext context) {
    // This method is rerun every time setState is called, for instance as done
    // by the _incrementCounter method above.
    //
    // The Flutter framework has been optimized to make rerunning build methods
    // fast, so that you can just rebuild anything that needs updating rather
    // than having to individually change instances of widgets.
    return Scaffold(
      appBar: AppBar(
        // TRY THIS: Try changing the color here to a specific color (to
        // Colors.amber, perhaps?) and trigger a hot reload to see the AppBar
        // change color while the other colors stay the same.
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        // Here we take the value from the MyHomePage object that was created by
        // the App.build method, and use it to set our appbar title.
        title: Text(widget.title),
      ),
      body: Center(
        // Center is a layout widget. It takes a single child and positions it
        // in the middle of the parent.
        child: Column(
          // Column is also a layout widget. It takes a list of children and
          // arranges them vertically. By default, it sizes itself to fit its
          // children horizontally, and tries to be as tall as its parent.
          //
          // Column has various properties to control how it sizes itself and
          // how it positions its children. Here we use mainAxisAlignment to
          // center the children vertically; the main axis here is the vertical
          // axis because Columns are vertical (the cross axis would be
          // horizontal).
          //
          // TRY THIS: Invoke "debug painting" (choose the "Toggle Debug Paint"
          // action in the IDE, or press "p" in the console), to see the
          // wireframe for each widget.
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            const Text(
              'You have pushed the button this many times:',
            ),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _incrementCounter,
        tooltip: 'Increment',
        child: const Icon(Icons.add),
      ), // This trailing comma makes auto-formatting nicer for build methods.
    );
  }
}

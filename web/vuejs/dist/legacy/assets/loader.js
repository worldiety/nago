(function (global, factory) {
  if (typeof define === "function" && define.amd) {
    define([], factory);
  } else if (typeof exports !== "undefined") {
    factory();
  } else {
    var mod = {
      exports: {}
    };
    factory();
    global.loader = mod.exports;
  }
})(typeof globalThis !== "undefined" ? globalThis : typeof self !== "undefined" ? self : this, function () {
  "use strict";

  // this script loads either a legacy (polyfilled) or modern build based on the client's browser version,
  // using Bowser to detect the browser and comparing it against version thresholds.
  (function () {
    function loadJSON(url, callback) {
      // we do not use let or const in this script to perform the script on legacy browsers as well
      var xhr = new XMLHttpRequest();
      // we do not use fetch here to perform the request on legacy browsers as well,
      xhr.open('GET', url, true);
      xhr.onreadystatechange = function () {
        // ready state 4 is 'DONE'
        if (xhr.readyState === 4) {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              var _data = JSON.parse(xhr.responseText);
              callback(null, _data);
            } catch (e) {
              callback(e, null);
            }
          } else {
            callback(new Error('Could not load URL: ' + url), null);
          }
        }
      };
      try {
        xhr.send(null);
      } catch (e) {
        callback(e, null);
      }
    }
    function loadScript(src, type) {
      var script = document.createElement('script');
      script.src = src;
      script.defer = true;
      script.type = type;
      document.head.appendChild(script);
      console.debug("Loaded script '".concat(src, "' with type '").concat(type, "'"));
    }
    function loadLink(rel, href) {
      var link = document.createElement('link');
      link.rel = rel;
      link.href = href;
      document.head.appendChild(link);
      console.debug("Loaded link '".concat(rel, "' with href '").concat(href, "'"));
    }

    // eslint-disable-next-line
    var browser = window.bowser.getParser(window.navigator.userAgent);
    var info = browser.getBrowser();
    var name = info.name;
    var version = parseInt(info.version, 10);
    var thresholds = {
      'Chrome': 103,
      'Firefox': 102,
      'Edge': 103,
      'Safari': 15,
      'IE': Infinity,
      'Internet Explorer': Infinity,
      'Opera': 89
    };
    var isOutdated = thresholds[name] !== undefined && version < thresholds[name];
    var buildType = isOutdated ? 'legacy' : 'modern';
    console.info("Detected browser '".concat(name, "' with version '").concat(version, "'. Loading build '").concat(buildType, "'..."));
    try {
      loadJSON("/".concat(buildType, "/manifest.json"), function (err, manifest) {
        if (err) {
          console.error("Error while loading build manifest for build ".concat(buildType), err);
          return;
        }

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        var mainEntry = manifest['src/main.ts'];
        if (!mainEntry || !mainEntry.file) {
          console.error("No entry file in build manifest found for build ".concat(buildType));
          return;
        }
        if (isOutdated) {
          // manually load polyfills
          loadScript("/".concat(buildType, "/assets/polyfill.min.js"), 'text/javascript');
          loadScript("/".concat(buildType, "/assets/minified.js"), 'text/javascript');
          loadScript("/".concat(buildType, "/assets/runtime.js"), 'text/javascript');
        }

        // load build and stylesheets
        loadScript("/".concat(buildType, "/").concat(mainEntry.file), isOutdated ? 'text/javascript' : 'module');
        if (mainEntry.css) {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          mainEntry.css.forEach(cssFile => {
            loadLink('stylesheet', "/".concat(buildType, "/").concat(cssFile));
          });
        }
        loadLink('manifest', "/".concat(buildType, "/manifest.json"));
        console.info("Successfully loaded build '".concat(buildType, "' for browser '").concat(name, "'.\nVersion '").concat(version, "' is ").concat(isOutdated ? 'older' : 'newer', " than the threshold '").concat(thresholds[name], "'."));
      });
    } catch (err) {
      console.error('Error while initializing the build:', err);
    }
  })();
});

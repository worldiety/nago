package core

import (
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/pkg/blob/crypto"
	"go.wdy.de/nago/presentation/ora"
	"strings"
)

type Navigation interface {
	ForwardTo(id NavigationPath, values Values)
	BackwardTo(id NavigationPath, values Values)
	Back()
	ResetTo(id NavigationPath, values Values)
	Reload()

	// Open either launches the denoted application or opens the resources with the associated resource.
	// This incorporates APIs like
	//  - the HTML5 open function or
	//  - the win32 ShellExecute call or
	//  - the MacOS open command or
	//  - xdg-open or gnome-open on linux
	//
	// This can be also used for other frontend primitives like handling oauth flows. See also [HTTPFlow].
	Open(resource URI, options Values)
}

type navigationController struct {
	destroyed bool
	scope     *Scope
}

func newNavigationController(scope *Scope) *navigationController {
	return &navigationController{
		scope: scope,
	}
}

func (n *navigationController) Window() Window {
	if n.scope != nil {
		return n.scope.allocatedRootView.UnwrapOr(nil)
	}

	return nil
}

func (n *navigationController) ForwardTo(id NavigationPath, values Values) {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationForwardToRequested{
		Type:    ora.NavigationForwardToRequestedT,
		Factory: ora.ComponentFactoryId(id),
		Values:  values,
	})
}

func (n *navigationController) BackwardTo(id NavigationPath, values Values) {
	// TODO implement me in ora protocol
	n.ForwardTo(id, values)
}

func (n *navigationController) Back() {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationBackRequested{
		Type: ora.NavigationBackRequestedT,
	})
}

func (n *navigationController) ResetTo(id NavigationPath, values Values) {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationResetRequested{
		Type:    ora.NavigationResetRequestedT,
		Factory: ora.ComponentFactoryId(id),
		Values:  values,
	})
}

func (n *navigationController) Reload() {
	if n.destroyed {
		return
	}

	n.scope.Publish(ora.NavigationReloadRequested{
		Type: ora.NavigationReloadRequestedT,
	})
}

func (n *navigationController) Open(resource URI, options Values) {
	n.scope.Publish(ora.OpenRequested{
		Type:     ora.OpenRequestedT,
		Resource: string(resource),
		Options:  options,
	})
}

// HTTPFlow issues either a WebView or a native browser redirect flow, starting at the given start-uri and in case of a
// result, expects the given redirectTarget.
// Example flow for a mobile app frontend:
//   - Android App opens a WebView with start uri
//   - If webview detects redirectTarget uri, it closes the webview AND
//   - extracts the query parameters AND
//   - invokes the navigation path with the query parameters
//
// Example for a web frontend:
//   - browser navigates to start
//   - browser redirects to redirectTarget which must be actually the redirectNavigation which we cannot intercept
//   - thus, the redirectNavigation is actually ignored for web frontends, however should be defined for completeness
//     for other frontends
//
// This function just uses [Navigation.Open] with the start as resource and the following defined options:
//   - _type is hardcoded as "http-flow"
//   - redirectTarget as declared
//   - redirectNavigation as declared
func HTTPFlow(nav Navigation, start, redirectTarget URI, redirectNavigation NavigationPath) {
	var encSid string
	if wnd, ok := nav.(interface{ Window() Window }); ok && wnd != nil {
		sid := wnd.Window().Session().ID()
		buf, err := crypto.Encrypt([]byte(sid), wnd.Window().Application().MasterKey())
		if err != nil {
			panic(fmt.Errorf("unreachable: %w", err))
		}

		encSid = hex.EncodeToString(buf)
	}

	nav.Open(start, Values{
		"_type":              "http-flow",
		"redirectTarget":     string(redirectTarget),
		"redirectNavigation": string(redirectNavigation),
		"session":            encSid,
	})
}

// HTTPify inspects the given string and eventually prefixes it with https://
func HTTPify(s string) URI {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return URI(s)
	}

	return "https://" + URI(s)
}

// HTTPOpen just triggers a regular (p)open call for the given http URL. A webfrontend will
// most likely trigger a window.open(url, target). In Javascript, target may be _blank|_self|_parent|_top|_unfencedTop
func HTTPOpen(nav Navigation, url URI, target string) {
	nav.Open(url, Values{
		"_type":  "http-link",
		"target": target,
	})
}

// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package testing

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/worldiety/option"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/std"
	"go.wdy.de/nago/presentation/proto"
	"log"
	"net/url"
	"reflect"
	"time"
)

type FindsCountMatcher struct {
	Amount int
}

func NewFindsCountMatcher(amount int) *FindsCountMatcher {
	return &FindsCountMatcher{
		Amount: amount,
	}
}

func (fcm *FindsCountMatcher) findsOneComponent() bool {
	return fcm.Amount == 1
}

func (fcm *FindsCountMatcher) findsNothing() bool {
	return fcm.Amount == 0
}

func (fcm *FindsCountMatcher) findsExactly(amount int) bool {
	return fcm.Amount == amount
}

func (fcm *FindsCountMatcher) findsAny() bool {
	return fcm.Amount >= 1
}

// Bestimmte Antwortzeit darf nicht überschritten werden
// Einen Button klicken
// Bestimmte Komponenten finden

type Tester struct {
}

func NewTester() *Tester {
	return &Tester{}
}

func (t *Tester) establishConnection(host string, path string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: path, RawQuery: "_sid=" + data.RandIdent[string]()}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	evt := proto.ScopeConfigurationChangeRequested{

		RID:            1,
		AcceptLanguage: "de",
		WindowInfo: proto.WindowInfo{
			Width:       800,
			Height:      600,
			Density:     2,
			SizeClass:   proto.SizeClassMedium,
			ColorScheme: proto.Light,
		},
	}

	var buf bytes.Buffer
	option.MustZero(proto.Marshal(proto.NewBinaryWriter(&buf), &evt))
	if err := c.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
		return nil, err
	}

	_, m, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}

	resp := std.Must(proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(m))))
	fmt.Println(resp.(*proto.ScopeConfigurationChanged).ApplicationID)

	//
	buf.Reset()
	option.MustZero(proto.Marshal(proto.NewBinaryWriter(&buf), &proto.SessionAssigned{SessionID: "1234"}))
	if err := c.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
		return nil, err
	}
	//
	return c, nil
}

func (t *Tester) tap(view proto.Component, text proto.Str) {

}

func (t *Tester) findByType(view proto.Component, typeToFind proto.Component) *FindsCountMatcher {
	var _, amount = findByType(view, typeToFind, 0)
	return NewFindsCountMatcher(amount)
}

func findByType(view proto.Component, typeToFind proto.Component, amount int) (bool, int) {
	if reflect.TypeOf(view) == reflect.TypeOf(typeToFind) {
		return true, amount + 1
	}

	switch view.(type) {
	case *proto.HStack:
		var hStack = view.(*proto.HStack)
		var amounts = 0
		var founds = false
		for _, c := range hStack.Children {
			var found, am = findByType(c, typeToFind, amount)
			if !founds && found {
				founds = true
			}
			if found {
				amounts += am
			}
		}
		if founds {
			return true, amounts
		}
		break
	case *proto.VStack:
		var vStack = view.(*proto.VStack)
		var amounts = 0
		var founds = false
		for _, c := range vStack.Children {
			var found, am = findByType(c, typeToFind, amount)
			if !founds && found {
				founds = true
			}
			if found {
				amounts += am
			}
		}
		if founds {
			return true, amounts
		}
		break
	}
	return false, amount
}

func (t *Tester) findText(view proto.Component, text proto.Str) bool {
	switch view.(type) {
	case *proto.Box:
	case *proto.Checkbox:
	case *proto.DatePicker:
	case *proto.Divider:
	case *proto.Grid:
	case *proto.HStack:
		var hStack = view.(*proto.HStack)
		for _, c := range hStack.Children {
			var found = t.findText(c, text)
			if found {
				return true
			}
		}
		break
	case *proto.Img:
	case *proto.Modal:
	case *proto.PasswordField:
	case *proto.Radiobutton:
	case *proto.Scaffold:
	case *proto.ScrollView:
	case *proto.Spacer:
	case *proto.Table:
	case *proto.TextField:
	case *proto.TextLayout:
	case *proto.TextView:
		var textView = view.(*proto.TextView)
		if textView.Value == text {
			return true
		}
	case *proto.Toggle:
	case *proto.VStack:
		var vStack = view.(*proto.VStack)
		for _, c := range vStack.Children {
			var found = t.findText(c, text)
			if found {
				return true
			}
		}
		break
	case *proto.WebView:
	case *proto.WindowTitle:
	}
	return false
}

func (t *Tester) renderViewWithMaxDuration(con *websocket.Conn, maxDurationInSec float64) (proto.Component, error, bool) {
	var startTime = time.Now()
	var endTime = time.Now()

	var resp, err = t.renderView(con)
	return resp, err, maxDurationInSec < endTime.Sub(startTime).Seconds()
}

func (t *Tester) renderView(con *websocket.Conn) (proto.Component, error) {

	var buf bytes.Buffer
	option.MustZero(proto.Marshal(proto.NewBinaryWriter(&buf), &proto.RootViewAllocationRequested{
		Locale:  "de",
		Factory: ".",
		RID:     2,
	}))

	if err := con.WriteMessage(websocket.BinaryMessage, buf.Bytes()); err != nil {
		return nil, err
	}

	//
	_, m, err := con.ReadMessage()
	if err != nil {
		return nil, err
	}
	resp := std.Must(proto.Unmarshal(proto.NewBinaryReader(bytes.NewBuffer(m))))

	return resp.(*proto.RootViewInvalidated).Root, nil
}

func (t *Tester) Test() error {
	for i := 0; i < 500; i++ {
		fmt.Println(i)

		con, conError := t.establishConnection("localhost:3000", "/wire")
		if conError != nil {
			fmt.Println(conError)
		}

		go func() {
			var rootView, err, reached = t.renderViewWithMaxDuration(con, 1)
			if err != nil {
				fmt.Println(err)
				return
			}

			if reached {
				log.Fatalf("The answer was too long")
			}

			var foundText = t.findText(rootView, "tertiary button")
			if !foundText {
				log.Fatalf("Could not find text: tertiary button!")
			}

			var typeToFind *proto.PasswordField
			var findsCountMatcher = t.findByType(rootView, typeToFind)
			if !findsCountMatcher.findsExactly(2) {
				log.Fatalf("Could not find 2 password fields")
			}

			var typeToFind2 *proto.Checkbox
			var findsCountMatcher2 = t.findByType(rootView, typeToFind2)
			if !findsCountMatcher2.findsNothing() {
				log.Fatalf("Has found a checkbox")
			}

			var typeToFind3 *proto.Toggle
			var findsCountMatcher3 = t.findByType(rootView, typeToFind3)
			if !findsCountMatcher3.findsOneComponent() {
				log.Fatalf("Could not find toggle")
			}

			//slog.Info("response is", "msg", resp)
		}()
	}

	return nil
}

package crud

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/ui"
	"log/slog"
	"strings"
)

func Text[E any, T ~string](label string, property func(*E) *T) Field[E] {
	return Field[E]{
		Label: label,
		RenderFormElement: func(self Field[E], entity *E) ui.DecoredView {
			state := core.StateOf[string](self.Window, self.ID).From(func() string {
				return string(*property(entity))
			})

			errState := core.StateOf[string](self.Window, self.ID+".err")

			state.Observe(func(newValue string) {
				f := property(entity)
				*f = T(newValue)
				if self.Validate != nil {
					errText, err := self.Validate(*entity)
					if err != nil {
						if errText == "" {
							var tmp [16]byte
							if _, err := rand.Read(tmp[:]); err != nil {
								panic(err)
							}
							incidentToken := hex.EncodeToString(tmp[:])
							errText = fmt.Sprintf("Unerwarteter Infrastrukturfehler: %s", incidentToken)
						}

						slog.Error(errText, "err", err)
					}

					errState.Set(errText)
				}

			})

			return ui.TextField(label, state.String()).
				InputValue(state).
				SupportingText(self.SupportingText).
				ErrorText(errState.Get())
		},
		RenderTableCell: func(self Field[E], entity E) ui.TTableCell {
			v := *property(&entity)
			return ui.TableCell(ui.Text(string(v)))
		},
		Comparator: func(a, b E) int {
			av := *property(&a)
			bv := *property(&b)
			return strings.Compare(string(av), string(bv))
		},
		Stringer: func(e E) string {
			return string(*property(&e))
		},
	}
}

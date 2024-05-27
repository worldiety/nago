package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"log/slog"
)

type Options[E any] struct {
	Title              string
	Create             func(E) error
	FindAll            iter.Seq2[E, error]
	Update             func(E) error
	AggregateActions   []AggregateAction[E]
	ForeignKeyMappings []ForeignKeyMapping
	Binding            *Binding[E]
}

func NewOptions[E any](with func(opts *Options[E])) *Options[E] {
	o := &Options[E]{}
	if with != nil {
		with(o)
	}

	return o
}

func (o *Options[E]) Delete(f func(E) error) {
	o.AggregateActions = append(o.AggregateActions, AggregateAction[E]{
		Icon:    icon.Trash,
		Caption: "",
		Action: func(owner ui.ModalOwner, e E) error {
			xdialog.ShowDelete(owner, "Soll der Eintrag wirklich gelöscht werden?", func() {
				if err := f(e); err != nil {
					xdialog.HandleError(owner, "Beim Löschen ist ein Fehler aufgetreten.", err)
				}
			}, nil)
			return nil
		},
		Style: ora.Destructive,
	})
}

type AggregateAction[T any] struct {
	Icon    ui.SVGSrc
	Caption string
	Action  func(ui.ModalOwner, T) error
	Style   ora.Intent
}

type ForeignKeyMapping struct {
}

func NewForeignKeyMapping[E data.Aggregate[ID], ID data.IDType]() *ForeignKeyMapping {
	return nil
}

func NewView[E any](owner ui.ModalOwner, opts *Options[E]) core.Component {
	if opts == nil {
		opts = NewOptions[E](nil)
	}

	if opts.Binding == nil {
		panic(fmt.Errorf("reflection based binder not yet implemented, please provide a custom binding"))
	}

	toolbar := ui.NewHStack(func(hstack *ui.FlexContainer) {
		hstack.ContentAlignment().Set(ora.ContentBetween)
		// left side
		hstack.Append(ui.MakeText(opts.Title))

		// right side
		hstack.Append(ui.NewHStack(func(hstack *ui.FlexContainer) {
			canSearch := opts.FindAll != nil
			if canSearch {
				hstack.ItemsAlignment().Set(ora.ItemsEnd)
				hstack.Append(ui.NewButton(func(btn *ui.Button) {
					btn.PreIcon().Set(icon.MagnifyingGlass)
					btn.Style().Set(ora.Tertiary)
				}))
				hstack.Append(ui.NewTextField(func(textField *ui.TextField) {
					textField.Placeholder().Set("Suchen")
					textField.Simple().Set(true)
				}))
			}

			if opts.Create != nil {
				hstack.Append(ui.NewButton(func(btn *ui.Button) {
					btn.Caption().Set("Neuer Eintrag")
					btn.PreIcon().Set(icon.Plus)
					btn.Action().Set(func() {
						ui.NewDialog(func(dlg *ui.Dialog) {
							dlg.Title().Set("Neuer Eintrag")
							dlg.Body().Set(opts.Binding.Form())
							dlg.Footer().Set(ui.NewHStack(func(hstack *ui.FlexContainer) {
								ui.HStackAlignRight(hstack)
								hstack.Append(ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("Abbrechen")
									btn.Style().Set(ora.Secondary)
									btn.Action().Set(func() {
										owner.Modals().Remove(dlg)
									})
								}))
								hstack.Append(ui.NewButton(func(btn *ui.Button) {
									btn.Caption().Set("Erstellen")
									btn.Action().Set(func() {
										var model E
										for _, field := range opts.Binding.fields {
											m, err := field.Form.IntoModel(model)
											if err != nil {
												field.Form.SetError(err.Error())
											} else {
												field.Form.SetError("")
											}

											model = m
										}

										if err := opts.Create(model); err != nil {
											xdialog.HandleError(owner, "Erstellen nicht möglich.", err)
										} else {
											owner.Modals().Remove(dlg)
										}

									})
								}))
							}))
							owner.Modals().Append(dlg)
						})
					})
				}))
			}
		}))
	})

	hasAggregateOptions := len(opts.AggregateActions) > 0

	return ui.NewVStack(func(vstack *ui.FlexContainer) {
		vstack.Append(toolbar)
		vstack.Append(ui.NewTable(func(table *ui.Table) {
			findAll := opts.FindAll
			if findAll == nil {
				slog.Info("cannot build table, FindAll iter is nil")
				return
			}

			for _, field := range opts.Binding.fields {
				table.Header().Append(ui.NewTextCell(field.Caption))
			}

			if hasAggregateOptions {
				table.Header().Append(ui.NewTextCell("Aktionen"))
			}

			table.Rows().From(func(yield func(*ui.TableRow) bool) {
				findAll(func(e E, err error) bool {
					if err != nil {
						slog.Error("cannot find entries", "err", err)
						return false
					}

					yield(ui.NewTableRow(func(row *ui.TableRow) {
						for _, field := range opts.Binding.fields {
							row.Cells().Append(ui.NewTextCell(field.Stringer(e)))
						}

						row.Cells().Append(ui.NewTableCell(func(cell *ui.TableCell) {
							cell.Body().Set(ui.NewHStack(func(hstack *ui.FlexContainer) {
								ui.HStackAlignRight(hstack)
								for _, action := range opts.AggregateActions {
									hstack.Append(ui.NewButton(func(btn *ui.Button) {
										btn.Caption().Set(action.Caption)
										btn.PreIcon().Set(action.Icon)
										btn.Style().Set(action.Style)
										btn.Action().Set(func() {
											xdialog.HandleError(owner, fmt.Sprintf("Aktion '%s' nicht durchführbar.", action.Caption), action.Action(owner, e))
										})
									}))
								}
							}))
						}))

					}))

					return true
				})
			})

		}))
	})
}

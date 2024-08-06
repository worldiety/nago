package xtable

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/iter"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/xdialog"
)

type SettingsID string

// deprecated: use crud package
type Settings struct {
	ID              SettingsID
	LastQuery       string
	SortByColumName string
	SortAsc         bool
}

func (s Settings) Identity() SettingsID {
	return s.ID
}

type AggregateAction[T any] struct {
	Icon    uilegacy.SVGSrc
	Caption string
	Action  func(T) error
	make    func(modals uilegacy.ModalOwner, t T) core.View
}

func (a AggregateAction[T]) makeComponent(modals uilegacy.ModalOwner, t T) core.View {
	if a.make == nil {
		return uilegacy.NewButton(func(btn *uilegacy.Button) {
			btn.Caption().Set(a.Caption)
			if a.Icon != "" {
				btn.PreIcon().Set(a.Icon)
			}
			btn.Action().Set(func() {
				//TODO i18n
				xdialog.HandleError(modals, fmt.Sprintf("Die Aktion '%s' hat einen Fehler ausgelöst.", a.Caption), a.Action(t))
			})
		})
	}

	return a.make(modals, t)
}

// deprecated: use crud package
// NewEditAction dispatches a standard action for editing to the given callback.
func NewEditAction[T any](onEdit func(T) error) AggregateAction[T] {
	return AggregateAction[T]{
		make: func(modals uilegacy.ModalOwner, t T) core.View {
			return uilegacy.NewButton(func(btn *uilegacy.Button) {
				btn.PreIcon().Set(icon.Pencil)
				btn.Style().Set(uilegacy.PrimaryIntent)
				btn.Action().Set(func() {
					//TODO i18n
					// usually does not happen, but who knows
					xdialog.HandleError(modals, "beim Bearbeiten ist ein Fehler aufgetreten", onEdit(t))
				})
			})
		},
	}
}

// NewDeleteAction returns a ready-to-use action which just removes the aggregate from the repository.
func NewDeleteAction[T any](delFn func(T) error) AggregateAction[T] {
	return AggregateAction[T]{
		make: func(modals uilegacy.ModalOwner, t T) core.View {
			return uilegacy.NewButton(func(btn *uilegacy.Button) {
				//TODO i18n
				btn.PreIcon().Set(icon.Trash)
				btn.Style().Set(uilegacy.Destructive)
				btn.Action().Set(func() {
					xdialog.ShowDelete(modals, "Soll der Eintrag wirklich unwiderruflich gelöscht werden?", func() {
						xdialog.HandleError(modals, "Beim Löschen ist ein Fehler aufgetreten.", delFn(t))
					}, nil)
				})
			})
		},
	}
}

// deprecated: use crud package
type Options[T any] struct {
	InstanceID       SettingsID
	Settings         data.Repository[Settings, SettingsID]
	CanSearch        bool
	PageSize         int
	AggregateActions []AggregateAction[T] // AggregateActions e.g. for editing (see [NewEditAction]) or delete (see [NewDeleteAction]) or something custom.
	Actions          []core.View          // Action buttons are used for table specific actions
}

// deprecated: use crud package
// NewTable creates a new simple data table view based on a repository.
func NewTable[T any](modals uilegacy.ModalOwner, items iter.Seq2[T, error], binding *Binding[T], opts Options[T]) core.View {
	if opts.PageSize == 0 {
		opts.PageSize = 20 // TODO: does that make sense for mobile at all?
	}

	hasEditColumn := len(opts.AggregateActions) > 0
	var settings Settings
	if opts.InstanceID != "" && opts.Settings != nil {
		optSettings, err := opts.Settings.FindByID(opts.InstanceID)
		xdialog.HandleError(modals, "cannot load table settings", err)
		if optSettings.Valid {
			settings = optSettings.V
		}
	}

	return uilegacy.NewVBox(func(vbox *uilegacy.VBox) {
		if opts.CanSearch {
			vbox.Append(
				uilegacy.NewTextField(func(searchField *uilegacy.TextField) {
					searchField.Label().Set("Filtern nach Stichworten")
					searchField.OnTextChanged().Set(func() {
						settings.LastQuery = searchField.Value().Get()
					})
				}),
			)
		}

		if len(opts.Actions) > 0 {
			vbox.Append(uilegacy.NewHBox(func(hbox *uilegacy.HBox) {
				hbox.Alignment().Set("flex-right") // TODO this is to web-centric
				for _, action := range opts.Actions {
					hbox.Append(action)
				}
			}))
		}

		var allSortBtns []*uilegacy.Button

		vbox.Append(
			uilegacy.NewTable(func(table *uilegacy.Table) {
				for _, col := range binding.Columns {
					table.Header().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
						if col.Sortable {
							cell.Body().Set(uilegacy.NewButton(func(btn *uilegacy.Button) {
								allSortBtns = append(allSortBtns, btn)
								btn.Caption().Set(col.Caption)
								btn.Style().Set(ora.Tertiary)
								btn.PreIcon().Set(icon.ArrowUpDown)
								btn.Action().Set(func() {
									for _, sortBtn := range allSortBtns {
										sortBtn.PreIcon().Set(icon.ArrowUpDown)
									}

									settings.SortByColumName = col.Caption
									settings.SortAsc = !settings.SortAsc
									if settings.SortAsc {
										btn.PreIcon().Set(icon.ArrowUp)
									} else {
										btn.PreIcon().Set(icon.ArrowDown)
									}

								})
							}))
						} else {
							cell.Body().Set(uilegacy.NewStr(col.Caption))
						}
					}))
				}

				if hasEditColumn {
					table.Header().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
						cell.Body().Set(uilegacy.NewStr("Optionen")) // todo i18n
					}))
				}

				table.Rows().From(func(yield func(*uilegacy.TableRow) bool) {
					rows, err := getData(items, binding, settings)
					if err != nil {
						vbox.Append(uilegacy.NewStr(fmt.Sprintf("error: %v", err)))
						return
					}

					for _, rowDat := range rows {
						yield(uilegacy.NewTableRow(func(row *uilegacy.TableRow) {
							for _, colText := range rowDat.values {
								row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
									cell.Body().Set(uilegacy.NewStr(colText))
								}))
							}

							if hasEditColumn {
								row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
									cell.Body().Set(uilegacy.NewHBox(func(hbox *uilegacy.HBox) {
										for _, action := range opts.AggregateActions {
											hbox.Append(action.makeComponent(modals, rowDat.model))
										}
									}))
								}))

							}
						}))
					}

				})
			}),
		)
	})
}

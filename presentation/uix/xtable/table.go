package xtable

import (
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"reflect"
)

type SettingsID string

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
	Icon    ui.SVGSrc
	Caption string
	Action  func(T) error
	make    func(modals ui.ModalOwner, t T) ui.LiveComponent
}

func (a AggregateAction[T]) makeComponent(modals ui.ModalOwner, t T) ui.LiveComponent {
	if a.make == nil {
		return ui.NewButton(func(btn *ui.Button) {
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

func NewEditAction[T any](onEdit func(T) error) AggregateAction[T] {
	return AggregateAction[T]{
		make: func(modals ui.ModalOwner, t T) ui.LiveComponent {
			return ui.NewButton(func(btn *ui.Button) {
				btn.PreIcon().Set(icon.Pencil)
				btn.Style().Set(ui.PrimaryIntent)
				btn.Action().Set(func() {
					//TODO i18n
					// usually does not happen, but who knows
					xdialog.HandleError(modals, "beim Bearbeiten ist ein Fehler aufgetreten", onEdit(t))
				})
			})
		},
	}
}

func NewDeleteAction[T any](onDelete func(T) error) AggregateAction[T] {
	return AggregateAction[T]{
		make: func(modals ui.ModalOwner, t T) ui.LiveComponent {
			return ui.NewButton(func(btn *ui.Button) {
				//TODO i18n
				btn.PreIcon().Set(icon.Trash)
				btn.Style().Set(ui.Destructive)
				btn.Action().Set(func() {
					xdialog.ShowDelete(modals, "Soll der Eintrag wirklich unwiderruflich gelöscht werden?", func() {
						xdialog.HandleError(modals, "Beim Löschen ist ein Fehler aufgetreten.", onDelete(t))
					}, nil)
				})
			})
		},
	}
}

type Options[E data.Aggregate[ID], ID data.IDType] struct {
	InstanceID       SettingsID
	Settings         data.Repository[Settings, SettingsID]
	CanSearch        bool
	CanSort          bool
	PageSize         int
	AggregateActions []AggregateAction[E] // AggregateActions e.g. for editing (see [NewEditAction]) or delete (see [NewDeleteAction]) or something custom.
	Actions          []ui.LiveComponent   // Action buttons are used for table specific actions
}

// NewTable creates a new simple data table view based on a repository. The ColumnModel can provide custom column names using a "caption" field tag.
func NewTable[E data.Aggregate[ID], ID data.IDType, ColumnModel any](modals ui.ModalOwner, repo data.Repository[E, ID], intoModel MapF[E, ColumnModel], opts Options[E, ID]) ui.LiveComponent {
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

	return ui.NewVBox(func(vbox *ui.VBox) {
		if opts.CanSearch {
			vbox.Append(
				ui.NewTextField(func(searchField *ui.TextField) {
					searchField.Label().Set("Filtern nach Stichworten")
					searchField.OnTextChanged().Set(func() {
						// todo trigger search
						settings.LastQuery = searchField.Value().Get()
					})
				}),
			)
		}

		if len(opts.Actions) > 0 {
			vbox.Append(ui.NewHBox(func(hbox *ui.HBox) {
				hbox.Alignment().Set("flex-right") // TODO this is to web-centric
				for _, action := range opts.Actions {
					hbox.Append(action)
				}
			}))
		}

		var zeroRow ColumnModel
		cols := getCols(zeroRow)

		var allSortBtns []*ui.Button

		vbox.Append(
			ui.NewTable(func(table *ui.Table) {
				for _, col := range cols {
					table.Header().Append(ui.NewTableCell(func(cell *ui.TableCell) {
						if opts.CanSort {
							cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
								allSortBtns = append(allSortBtns, btn)
								btn.Caption().Set(col.name)
								btn.PreIcon().Set(icon.ArrowUpDown)
								btn.Action().Set(func() {
									for _, sortBtn := range allSortBtns {
										sortBtn.PreIcon().Set(icon.ArrowUpDown)
									}

									settings.SortByColumName = col.name
									settings.SortAsc = !settings.SortAsc
									if settings.SortAsc {
										btn.PreIcon().Set(icon.ArrowUp)
									} else {
										btn.PreIcon().Set(icon.ArrowDown)
									}

								})
							}))
						} else {
							cell.Body().Set(ui.MakeText(col.name))
						}
					}))
				}

				if hasEditColumn {
					table.Header().Append(ui.NewTableCell(func(cell *ui.TableCell) {
						cell.Body().Set(ui.MakeText("Optionen")) // todo i18n
					}))
				}

				table.Rows().From(func(yield func(*ui.TableRow)) {
					rows, err := getData(repo, intoModel, opts, settings)
					if err != nil {
						vbox.Append(ui.MakeText(fmt.Sprintf("error: %v", err)))
						return // TODO wrong seq signature
					}

					for _, rowDat := range rows {
						yield(ui.NewTableRow(func(row *ui.TableRow) {
							for _, col := range rowDat.cols {
								row.Cells().Append(ui.NewTableCell(func(cell *ui.TableCell) {
									cell.Body().Set(ui.MakeText(col.value))
								}))
							}

							if hasEditColumn {
								row.Cells().Append(ui.NewTableCell(func(cell *ui.TableCell) {
									cell.Body().Set(ui.NewHBox(func(hbox *ui.HBox) {
										for _, action := range opts.AggregateActions {
											hbox.Append(action.makeComponent(modals, rowDat.holder.Original))
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

type column struct {
	name  string
	field reflect.StructField
}

func getCols(t any) []column {
	var res []column
	rType := reflect.TypeOf(t)
	for i := range rType.NumField() {
		field := rType.Field(i)
		if !field.IsExported() {
			continue
		}

		caption, ok := field.Tag.Lookup("caption")
		if !ok {
			caption = field.Name
		}

		res = append(res, column{
			name:  caption,
			field: field,
		})
	}

	return res
}

func getColData(t any) []colData {
	var res []colData
	rType := reflect.TypeOf(t)
	rVal := reflect.ValueOf(t)
	for i := range rType.NumField() {
		field := rType.Field(i)
		if !field.IsExported() {
			continue
		}

		res = append(res, colData{value: fmt.Sprintf("%v", rVal.Field(i).Interface())})
	}

	return res
}

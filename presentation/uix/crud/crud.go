package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/iter"
	slices2 "go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"log/slog"
	"slices"
	"strings"
)

type Options[E any] struct {
	title            string
	actions          []core.Component // global components to show for the entire crud set, e.g. for custom create action
	create           func(E) error
	prepareCreate    func(E) (E, error)
	findAll          iter.Seq2[E, error]
	aggregateActions []AggregateAction[E]
	binding          *Binding[E]
}

func NewOptions[E any](with func(opts *Options[E])) *Options[E] {
	o := &Options[E]{}
	if with != nil {
		with(o)
	}

	return o
}

// Actions adds the given components into the CRUD global action area.
func (o *Options[E]) Actions(actions ...core.Component) {
	o.actions = append(o.actions, actions...)
}

func (o *Options[E]) Title(s string) *Options[E] {
	o.title = s
	return o
}

func (o *Options[E]) PrepareCreate(f func(E) (E, error)) *Options[E] {
	o.prepareCreate = f
	return o
}

func (o *Options[E]) Create(f func(E) error) *Options[E] {
	o.create = f
	return o
}

func (o *Options[E]) ReadAll(it iter.Seq2[E, error]) *Options[E] {
	o.findAll = it
	return o
}

func (o *Options[E]) Delete(f func(E) error) *Options[E] {
	o.aggregateActions = append(o.aggregateActions, AggregateAction[E]{
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

	return o
}

// Bind allocates a new explicit data binding and sets it into the options.
func (o *Options[E]) Bind(with func(binding *Binding[E])) *Options[E] {
	o.binding = NewBinding[E](with)
	return o
}

func (o *Options[E]) Binding(binding *Binding[E]) *Options[E] {
	o.binding = binding
	return o
}

func (o *Options[E]) Update(f func(E) error) *Options[E] {
	opts := o
	o.aggregateActions = append(o.aggregateActions, AggregateAction[E]{
		Icon: icon.Pencil,
		Action: func(owner ui.ModalOwner, e E) error {
			ui.NewDialog(func(dlg *ui.Dialog) {
				dlg.Title().Set("Eintrag bearbeiten")
				dlg.Size().Set(ora.ElementSizeMedium)
				form := opts.binding.NewForm(Update)
				dlg.Body().Set(form.Component)
				// push actual model data into the view
				for _, field := range form.Fields {
					field.FromModel(e)
				}

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
						btn.Caption().Set("Aktualisieren")
						btn.Action().Set(func() {
							for _, field := range form.Fields {
								m, err := field.IntoModel(e)
								if err != nil {
									field.SetError(err.Error())
								} else {
									field.SetError("")
								}

								e = m
							}

							if err := f(e); err != nil {
								xdialog.HandleError(owner, "Speichern nicht möglich.", err)
							} else {
								owner.Modals().Remove(dlg)
							}

						})
					}))
				}))
				owner.Modals().Append(dlg)
			})

			return nil
		},
		Style: ora.Primary,
	})

	return o
}

type AggregateAction[T any] struct {
	Icon    ui.SVGSrc
	Caption string
	Action  func(ui.ModalOwner, T) error
	Style   ora.Intent
}

func NewView[E any](owner ui.ModalOwner, opts *Options[E]) core.Component {
	if opts == nil {
		opts = NewOptions[E](nil)
	}

	if opts.binding == nil {
		panic(fmt.Errorf("reflection based binder not yet implemented, please provide a custom binding"))
	}
	var searchField *ui.TextField
	toolbar := ui.NewHStack(func(hstack *ui.FlexContainer) {
		hstack.ContentAlignment().Set(ora.ContentBetween)
		// left side
		hstack.Append(ui.MakeText(opts.title))

		// right side

		hstack.Append(ui.NewHStack(func(hstack *ui.FlexContainer) {
			canSearch := opts.findAll != nil
			if canSearch {
				hstack.ItemsAlignment().Set(ora.ItemsEnd)
				hstack.Append(ui.NewButton(func(btn *ui.Button) {
					btn.PreIcon().Set(icon.MagnifyingGlass)
					btn.Style().Set(ora.Tertiary)
				}))
				hstack.Append(ui.NewTextField(func(textField *ui.TextField) {
					searchField = textField
					textField.Placeholder().Set("Suchen")
					textField.Simple().Set(true)
				}))
			}

			for _, action := range opts.actions {
				hstack.Append(action)
			}

			if opts.create != nil {
				hstack.Append(ui.NewButton(func(btn *ui.Button) {
					btn.Caption().Set("Neuer Eintrag")
					btn.PreIcon().Set(icon.Plus)
					btn.Action().Set(func() {
						ui.NewDialog(func(dlg *ui.Dialog) {
							dlg.Title().Set("Neuer Eintrag")
							dlg.Size().Set(ora.ElementSizeMedium)
							form := opts.binding.NewForm(Create)
							dlg.Body().Set(form.Component)
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
										for _, field := range form.Fields {
											m, err := field.IntoModel(model)
											if err != nil {
												field.SetError(err.Error())
											} else {
												field.SetError("")
											}

											model = m
										}

										if opts.prepareCreate != nil {
											m, err := opts.prepareCreate(model)
											if err != nil {
												xdialog.HandleError(owner, "Vorbereiten zum Erstellen nicht möglich.", err)
											}
											model = m
										}
										if err := opts.create(model); err != nil {
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

	hasAggregateOptions := len(opts.aggregateActions) > 0

	return ui.NewVStack(func(vstack *ui.FlexContainer) {
		vstack.Append(toolbar)
		vstack.Append(ui.NewTable(func(table *ui.Table) {
			findAll := opts.findAll
			if findAll == nil {
				slog.Info("cannot build table, findAll iter is nil")
				return
			}

			var allSortBtns []*ui.Button
			sortAsc := true
			sortByFieldIdx := -1

			for i, field := range opts.binding.fields {
				if field.RenderHints[Overview] == Hidden {
					continue
				}
				table.Header().Append(ui.NewTableCell(func(cell *ui.TableCell) {
					cell.Body().Set(ui.NewButton(func(btn *ui.Button) {
						btn.Style().Set(ora.Tertiary)
						btn.Caption().Set(field.Caption)
						btn.PreIcon().Set(icon.ArrowUpDown)
						allSortBtns = append(allSortBtns, btn)
						btn.Action().Set(func() {
							sortByFieldIdx = i
							// reset that sort icon
							for _, sortBtn := range allSortBtns {
								sortBtn.PreIcon().Set(icon.ArrowUpDown)
							}

							sortAsc = !sortAsc
							if sortAsc {
								btn.PreIcon().Set(icon.ArrowUp)
							} else {
								btn.PreIcon().Set(icon.ArrowDown)
							}
						})
					}))
				}))
			}

			if hasAggregateOptions {
				table.Header().Append(ui.NewTextCell("Aktionen"))
			}

			table.Rows().From(func(yield func(*ui.TableRow) bool) {
				filtered := findAll
				if searchField.Value().Get() != "" {
					predicate := rquery.SimplePredicate[any](searchField.Value().Get())
					filtered = iter.Filter2(func(model E, err error) bool {
						if err != nil {
							slog.Error("error in iter while filtering", "err", err)
							return false
						}

						// note that this may be a security faux-pas, because we can search things which is not displayed,
						// thus an attacker may "leak" information through search responses. However, this works as intended
						// and allows to search after hidden but well known details, like internal entity identifiers.
						// To mitigate security problems, the developer just needs to use a proper view model,
						// as required by a reasonable architecture anyway.
						return predicate(model)
					}, findAll)
				}

				if sortByFieldIdx >= 0 {
					var err error
					tmpIter := iter.BreakOnError(&err, filtered)
					collectedRows := slices2.Collect(tmpIter)
					if err != nil {
						slog.Error("error in iter while collecting", "err", err)
						return
					}

					field := opts.binding.fields[sortByFieldIdx]
					slices.SortFunc(collectedRows, func(a, b E) int {
						strA := field.Stringer(a)
						strB := field.Stringer(b)
						dir := 1
						if !sortAsc {
							dir = -1
						}
						return strings.Compare(strA, strB) * dir
					})
					filtered = slices2.Values2[[]E, E, error](collectedRows)
				}

				filtered(func(e E, err error) bool {
					if err != nil {
						slog.Error("cannot find entries", "err", err)
						return false
					}

					yield(ui.NewTableRow(func(row *ui.TableRow) {
						for _, field := range opts.binding.fields {
							if field.RenderHints[Overview] == Hidden {
								continue
							}
							row.Cells().Append(ui.NewTextCell(field.Stringer(e)))
						}

						if len(opts.aggregateActions) > 0 {
							row.Cells().Append(ui.NewTableCell(func(cell *ui.TableCell) {
								cell.Body().Set(ui.NewHStack(func(hstack *ui.FlexContainer) {
									ui.HStackAlignRight(hstack)
									for _, action := range opts.aggregateActions {
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
						}

					}))

					return true
				})
			})

		}))
	})
}

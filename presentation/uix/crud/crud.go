package crud

import (
	"fmt"
	"go.wdy.de/nago/pkg/data/rquery"
	"go.wdy.de/nago/pkg/iter"
	slices2 "go.wdy.de/nago/pkg/slices"
	"go.wdy.de/nago/presentation/core"
	"go.wdy.de/nago/presentation/icon"
	"go.wdy.de/nago/presentation/ora"
	"go.wdy.de/nago/presentation/uilegacy"
	"go.wdy.de/nago/presentation/uix/xdialog"
	"log/slog"
	"slices"
	"strings"
)

type Options[E any] struct {
	title            string
	actions          []core.View // global components to show for the entire crud set, e.g. for custom create action
	create           func(E) error
	prepareCreate    func(E) (E, error)
	findAll          iter.Seq2[E, error]
	aggregateActions []AggregateAction[E]
	binding          *Binding[E]
	wnd              core.Window
}

func NewOptions[E any](with func(opts *Options[E])) *Options[E] {
	o := &Options[E]{}
	if with != nil {
		with(o)
	}

	return o
}

// Actions adds the given components into the CRUD global action area.
func (o *Options[E]) Actions(actions ...core.View) {
	o.actions = append(o.actions, actions...)
}

// AggregateActions adds the given actions to each individual entity or aggregate entry.
func (o *Options[E]) AggregateActions(actions ...AggregateAction[E]) *Options[E] {
	o.aggregateActions = append(o.aggregateActions, actions...)
	return o
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

func (o *Options[E]) Responsive(wnd core.Window) *Options[E] {
	o.wnd = wnd
	return o
}

type AggregeActionOption[E any] func(*AggregateAction[E])

func AggregationActionOptionVisibility[E any](f func(E) bool) AggregeActionOption[E] {
	return func(action *AggregateAction[E]) {
		action.visible = f
	}
}

func (o *Options[E]) Delete(f func(E) error, options ...AggregeActionOption[E]) *Options[E] {
	a := AggregateAction[E]{
		Icon:    icon.Trash,
		Caption: "",
		Action: func(owner uilegacy.ModalOwner, e E) error {
			xdialog.ShowDelete(owner, "Soll der Eintrag wirklich gelöscht werden?", func() {
				if err := f(e); err != nil {
					xdialog.HandleError(owner, "Beim Löschen ist ein Fehler aufgetreten.", err)
				}
			}, nil)
			return nil
		},
<<<<<<< HEAD
		Style: ora.Primary,
	})
=======
		Style: ora.Destructive,
	}.WithOptions(options...)

	o.aggregateActions = append(o.aggregateActions, a)
>>>>>>> 8898002c6c3b896349032d4f8b92f2318d75a45d

	return o
}

// Bind allocates a new explicit data binding and sets it into the options.
func (o *Options[E]) Bind(with func(bnd *Binding[E])) *Options[E] {
	o.binding = NewBinding[E](with)
	return o
}

func (o *Options[E]) Binding(binding *Binding[E]) *Options[E] {
	o.binding = binding
	return o
}

func (o *Options[E]) Update(f func(E) error, options ...AggregeActionOption[E]) *Options[E] {
	opts := o
	action := AggregateAction[E]{
		Icon: icon.Pencil,
		Action: func(owner uilegacy.ModalOwner, e E) error {
			uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
				dlg.Title().Set("Eintrag bearbeiten")
				dlg.Size().Set(ora.ElementSizeMedium)
				form := opts.binding.NewForm(Update)
				dlg.Body().Set(form.Component)
				// push actual model data into the view
				for _, field := range form.Fields {
					field.FromModel(e)
				}

				dlg.Footer().Set(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
					hstack.SetAlignment(ora.Leading)
					hstack.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
						btn.Caption().Set("Abbrechen")
						btn.Style().Set(ora.Secondary)
						btn.Action().Set(func() {
							owner.Modals().Remove(dlg)
						})
					}))
					hstack.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
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
	}.WithOptions(options...)

	o.aggregateActions = append(o.aggregateActions, action)

	return o
}

type AggregateAction[T any] struct {
	Icon    uilegacy.SVGSrc
	Caption string
	Action  func(uilegacy.ModalOwner, T) error
<<<<<<< HEAD
	Style   ora.Color
=======
	Style   ora.Intent
	visible func(T) bool
}

func (a AggregateAction[T]) WithOptions(options ...AggregeActionOption[T]) AggregateAction[T] {
	for _, opt := range options {
		opt(&a)
	}
	return a
>>>>>>> 8898002c6c3b896349032d4f8b92f2318d75a45d
}

func NewView[E any](owner uilegacy.ModalOwner, opts *Options[E]) core.View {
	if opts == nil {
		opts = NewOptions[E](nil)
	}

	if opts.binding == nil {
		panic(fmt.Errorf("reflection based binder not yet implemented, please provide a custom binding"))
	}

	var searchField *uilegacy.TextField
	toolbar := uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
		hstack.SetAlignment(ora.Trailing) // TODO this should be content between
		// left side
		hstack.Append(uilegacy.NewStr(opts.title))

		// right side

		hstack.Append(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
			canSearch := opts.findAll != nil
			if canSearch {
				hstack.SetAlignment(ora.Trailing)
				hstack.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
					btn.PreIcon().Set(icon.MagnifyingGlass)
					btn.Style().Set(ora.Tertiary)
				}))
				hstack.Append(uilegacy.NewTextField(func(textField *uilegacy.TextField) {
					searchField = textField
					textField.OnDebouncedTextChanged().Set(func() {
						// nothing to do, this will trigger invalidation anyway
					})
					textField.Placeholder().Set("Suchen")
					textField.Simple().Set(true)
				}))
			}

			for _, action := range opts.actions {
				hstack.Append(action)
			}

			if opts.create != nil {
				hstack.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
					btn.Caption().Set("Neuer Eintrag")
					btn.PreIcon().Set(icon.Plus)
					btn.Action().Set(func() {
						uilegacy.NewDialog(func(dlg *uilegacy.Dialog) {
							dlg.Title().Set("Neuer Eintrag")
							dlg.Size().Set(ora.ElementSizeMedium)
							form := opts.binding.NewForm(Create)
							dlg.Body().Set(form.Component)
							dlg.Footer().Set(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
								hstack.SetAlignment(ora.Trailing)
								hstack.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
									btn.Caption().Set("Abbrechen")
									btn.Style().Set(ora.Secondary)
									btn.Action().Set(func() {
										owner.Modals().Remove(dlg)
									})
								}))
								hstack.Append(uilegacy.NewButton(func(btn *uilegacy.Button) {
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
	var componentBody uilegacy.Container

	_ = toolbar
	setupExpandedTable := func() {
		componentBody.Children().Clear()
		componentBody.Children().Append(toolbar)
		componentBody.Children().Append(uilegacy.NewTable(func(table *uilegacy.Table) {
			findAll := opts.findAll
			if findAll == nil {
				slog.Info("cannot build table, findAll iter is nil")
				return
			}

			var allSortBtns []*uilegacy.Button
			sortAsc := true
			sortByFieldIdx := -1

			for i, field := range opts.binding.fields {
				if field.RenderHints[Overview] == Hidden {
					continue
				}
				table.Header().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
					cell.Body().Set(uilegacy.NewButton(func(btn *uilegacy.Button) {
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
				table.Header().Append(uilegacy.NewTextCell("Aktionen"))
			}

			table.Rows().From(func(yield func(*uilegacy.TableRow) bool) {
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
					if field.CompareField == nil {
						slices.SortFunc(collectedRows, func(a, b E) int {
							strA := field.Stringer(a)
							strB := field.Stringer(b)
							dir := 1
							if !sortAsc {
								dir = -1
							}
							return strings.Compare(strA, strB) * dir
						})
					} else {
						slices.SortFunc(collectedRows, func(a, b E) int {
							dir := 1
							if !sortAsc {
								dir = -1
							}
							return field.CompareField(a, b) * dir
						})
					}

					filtered = slices2.Values2[[]E, E, error](collectedRows)
				}

				filtered(func(e E, err error) bool {
					if err != nil {
						slog.Error("cannot find entries", "err", err)
						return false
					}

					yield(uilegacy.NewTableRow(func(row *uilegacy.TableRow) {
						for _, field := range opts.binding.fields {
							if field.RenderHints[Overview] == Hidden {
								continue
							}
							row.Cells().Append(uilegacy.NewTextCell(field.Stringer(e)))
						}

						if len(opts.aggregateActions) > 0 {
							row.Cells().Append(uilegacy.NewTableCell(func(cell *uilegacy.TableCell) {
								cell.Body().Set(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
									hstack.SetAlignment(ora.Trailing)
									for _, action := range opts.aggregateActions {
										hstack.Append(newAggregateActionButton(owner, action, e))
									}
								}))
							}))
						}

					}))

					return true
				})
			})

		}))
	}

	setupListDataAsCards := func() {
		componentBody.Children().Clear()
		componentBody.Children().Append(toolbar)
		componentBody.Children().Append(uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
			//vstack.ElementSize().Set(ora.ElementSizeLarge) // TODO what do we make with the element size large thingy here?
			vstack.Bla = func(yield func(core.View) bool) {
				findAll := opts.findAll
				if findAll == nil {
					slog.Info("cannot build table, findAll iter is nil")
					return
				}

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

				filtered(func(e E, err error) bool {
					if err != nil {
						slog.Error("error in iter", "err", err)
						yield(xdialog.ErrorView("iter failed", err))
						return false
					}

					return yield(uilegacy.NewCard(func(card *uilegacy.Card) {
						card.Append(uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
							for _, field := range opts.binding.fields {
								if field.RenderHints[Card] == Title {
									vstack.Append(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
										hstack.SetAlignment(ora.Trailing) // TODO this should be in-between
										hstack.Append(uilegacy.NewStr(field.Caption))
										hstack.Append(uilegacy.NewStr(field.Stringer(e)))
									}))
									vstack.Append(uilegacy.NewDivider(nil))
									break

								}
							}

							for _, field := range opts.binding.fields {
								if field.RenderHints[Overview] == Hidden {
									continue
								}

								if field.RenderHints[Card] == Title {
									continue
								}

								vstack.Append(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
									hstack.SetAlignment(ora.Trailing) // TODO this should be inbetween
									hstack.Append(uilegacy.NewText(func(t *uilegacy.Text) {
										t.Value().Set(field.Caption)
										t.Size().Set("lg")
									}))
									hstack.Append(uilegacy.NewStr(field.Stringer(e)))
								}))

							}

							if len(opts.aggregateActions) > 0 {
								vstack.Append(uilegacy.NewHStack(func(hstack *uilegacy.HStack) {
									hstack.SetAlignment(ora.Trailing)
									for _, action := range opts.aggregateActions {

										hstack.Append(newAggregateActionButton(owner, action, e))
									}
								}))
							}

						}))

					}))

				})
			}
		}))
	}

	sizeClass := ora.ExpandedWindow

	renderBody := func() {
		switch {
		case ora.ExpandedWindow != sizeClass:
			setupListDataAsCards()
		default:
			setupExpandedTable()
		}

	}

	if opts.wnd != nil {
		hasCardTitle := false
		for _, field := range opts.binding.fields {
			if field.RenderHints[Card] == Title {
				hasCardTitle = true
				break
			}
		}

		if !hasCardTitle && len(opts.binding.fields) > 0 {
			if opts.binding.fields[0].RenderHints == nil {
				opts.binding.fields[0].RenderHints = map[RenderVariant]RenderHint{}
			}
			opts.binding.fields[0].RenderHints[Card] = Title
		}

		sizeClass = opts.wnd.WindowInfo().SizeClass()
		opts.wnd.ViewRoot().AddWindowSizeClassObserver(func(size ora.WindowSizeClass) {
			sizeClass = size
			renderBody()
		})
	}

	return uilegacy.NewVStack(func(vstack *uilegacy.VStack) {
		//componentBody = vstack
		panic("fix me")
		renderBody()
	})
}

func newAggregateActionButton[E any](owner uilegacy.ModalOwner, action AggregateAction[E], e E) core.View {
	return uilegacy.NewButton(func(btn *uilegacy.Button) {
		if action.visible != nil {
			btn.Visible().Set(action.visible(e))
		}
		btn.Caption().Set(action.Caption)
		btn.PreIcon().Set(action.Icon)
		if action.Style != "" {
			btn.Style().Set(action.Style)
		}
		btn.Action().Set(func() {
			xdialog.HandleError(owner, fmt.Sprintf("Aktion '%s' nicht durchführbar.", action.Caption), action.Action(owner, e))
		})
	})
}

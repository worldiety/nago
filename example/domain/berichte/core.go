package berichte

type Prüfbericht struct {
	Prüfungen []Prüfung
}

type Prüfung struct {
	ID        string
	Bestanden bool
}

func PrüfungHinzufügen(repo any, bericht Prüfbericht, prüfung Prüfung) error {
	panic("???")
}

type WorkflowEngine interface {
	NewInstance() string
	SaveState(any)
}

func NewWorkflow(w WorkflowEngine) NeuenBerichtInBearbeitung {
	return NeuenBerichtInBearbeitung{Instance: w.NewInstance()}
}

type NeuenBerichtInBearbeitung struct {
	Instance string
	engine   WorkflowEngine
}

type NameGesetzt string

func (b NeuenBerichtInBearbeitung) PrüfungHinzufügen(name NameGesetzt) NeuenBerichtInBearbeitung {
	b.engine.SaveState(b)
	return b
}

type NutzerIstFertig string

func (b NeuenBerichtInBearbeitung) Fertigstellen(NutzerIstFertig) BerichtFertiggestellt {
	newState := BerichtFertiggestellt{}
	b.engine.SaveState(newState)
	return newState
}

type BerichtFertiggestellt struct {
}

func (BerichtFertiggestellt) CloseWorkflow() {

}

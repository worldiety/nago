package workflow_test

import (
	"go.wdy.de/nago/pkg/pubsub"
	"go.wdy.de/nago/pkg/workflow"
	"go.wdy.de/nago/pkg/xreflect"
	"testing"
)

type PizzaWurdeBestellt struct {
	PizzaName     string
	TelefonNummer string
	Anzahl        int
}

type PizzaBestellungAngenommen struct {
	Auftragsnummer string
}

func TestCreate(t *testing.T) {
	xreflect.SetTypeDoc(xreflect.TypeIDOf[PizzaWurdeBestellt](), "PizzaWurdeBestellt")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[PizzaBestellungAngenommen](), "PizzaBestellungAngenommen")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[PizzaWasEaten](), "PizzaWasEaten")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[OrderPizza](), "OrderPizza")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[CheckIfPizzaWasDelieveredAfter30](), "checkIfPizzaWasDelieveredAfter30")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[CheckIfPizzaWasDelieveredAfter50](), "checkIfPizzaWasDelieveredAfter50")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[CancelSubscriptionRequested](), "CancelSubscriptionRequested")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[SubscriptionCancelled](), "SubscriptionCancelled")
	xreflect.SetTypeDoc(xreflect.TypeIDOf[CancelSubscription](), "CancelSubscription")

	pubsub := pubsub.NewPubSub()
	func(orderPizza OrderPizza) {
		wf := workflow.Create(pubsub, "Pizza-Bestellung", 1, func(wf *workflow.Workflow) {
			workflow.Subscribe(wf, orderPizza)
			workflow.Subscribe(wf, CheckIfPizzaWasDelieveredAfter30(checkIfPizzaWasDelieveredAfter30))
			workflow.Subscribe(wf, CheckIfPizzaWasDelieveredAfter50(checkIfPizzaWasDelieveredAfter50))
			workflow.Subscribe(wf, CancelSubscription(nil))
		})

		instance := wf.NewInstance()
		instance.ID()

		t.Log(wf.String())
	}(newOrderPizza())

}

type CancelSubscription func(CancelSubscriptionRequested) SubscriptionCancelled

type OrderPizza func(PizzaWurdeBestellt) PizzaBestellungAngenommen

func newOrderPizza() OrderPizza {
	return func(bestellt PizzaWurdeBestellt) PizzaBestellungAngenommen {
		panic("blub")
	}
}

type CancelSubscriptionRequested struct {
	SubscriptionID string
}

type SubscriptionCancelled struct{}

type PizzaWasEaten struct{}

type CheckIfPizzaWasDelieveredAfter30 func(PizzaBestellungAngenommen) PizzaWasEaten

// TODO this has its own cursor even though its the same event? keep event + stage transit = trigger after start per instance?
func checkIfPizzaWasDelieveredAfter30(PizzaBestellungAngenommen) PizzaWasEaten {
	return PizzaWasEaten{}
}

// TODO this has its own cursor even though its the same event? keep event + stage transit = trigger after start per instance?
type CheckIfPizzaWasDelieveredAfter50 func(PizzaBestellungAngenommen) PizzaWasEaten

func checkIfPizzaWasDelieveredAfter50(PizzaBestellungAngenommen) PizzaWasEaten {
	return PizzaWasEaten{}
}

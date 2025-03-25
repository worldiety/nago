package xreflect

import (
	"testing"
)

type HelloWorld struct {
}

func (HelloWorld) HelloWorld(world HelloWorld) {

}

type MyUseCaseFn func(world HelloWorld)

type MyService interface {
	HelloWorld(world HelloWorld)
}

func TestTypeIDOf(t *testing.T) {
	t.Log(TypeIDOf[HelloWorld]())
	t.Log(TypeIDOf[*HelloWorld]())
	t.Log(TypeIDOf[MyUseCaseFn]())
	t.Log(TypeIDOf[func(world HelloWorld)]())
	t.Log(TypeIDOf[MyService]())
	t.Log(TypeIDOf[any]())
	//t.Log(TypeIDOf[icons.Test[MyService]]())

	t.Log("--")
	var a HelloWorld
	var b *HelloWorld
	var c MyUseCaseFn
	var d *MyService // note that var d MyService is a nil-interface-type, which is not allowed
	t.Log(TypeIDFrom(a))
	t.Log(TypeIDFrom(b))
	t.Log(TypeIDFrom(c))
	t.Log(TypeIDFrom(d))
	t.Log(TypeIDFrom(nil))
	t.Log(TypeIDFrom(a.HelloWorld))
}

func BenchmarkTypeIDOf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		TypeIDOf[HelloWorld]()
	}
}

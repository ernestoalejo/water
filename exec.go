package main

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
	zero      reflect.Value
)

type variables map[string]reflect.Value
type functions map[string]reflect.Value

type state struct {
	funcs  functions
	vars   variables
	output io.Writer
	t      *ListNode
}

func (s *state) recover(errp *error) {
	if e := recover(); e != nil {
		*errp = fmt.Errorf("%s", e)
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
	}
}

func (s *state) errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (s *state) exec() reflect.Value {
	for _, node := range s.t.Nodes {
		switch node.Type() {
		case NodeCall:
			return s.makeCall(node.(*CallNode))
		}
	}

	return zero
}

func (s *state) makeCall(n *CallNode) reflect.Value {
	// Get the func in the index
	f, ok := s.funcs[n.Name]
	if !ok {
		panic(fmt.Errorf("function not defined %s", n.Name))
	}

	// Analyze its type
	t := f.Type()
	numArgs := t.NumIn()

	// Check if the number of args it's correct
	if t.IsVariadic() {
		numArgs -= 1
		if len(n.Args) < numArgs {
			s.errorf("wrong number of args for %s: want at least %d, got %d", n.Name,
				numArgs, len(n.Args))
		}
	} else if len(n.Args) != numArgs {
		s.errorf("wrong number of args for %s: want %d, got %d", n.Name, numArgs, len(n.Args))
	}

	// Check if the function return it's correct
	s.checkFuncReturn(n.Name, t)

	// Prepare the arguments array
	args := make([]reflect.Value, numArgs)

	// Add the fixed arguments
	i := 0
	for ; i < numArgs; i++ {
		args[i] = s.evalArg(t.In(i), n.Args[i])
	}

	// Add the variadic arguments
	if t.IsVariadic() {
		argType := t.In(numArgs).Elem()
		for ; i < len(n.Args); i++ {
			args[i] = s.evalArg(argType, n.Args[i])
		}
	}

	res := f.Call(args)

	if len(res) == 2 && !res[1].IsNil() {
		s.errorf("error calling %s: %s", n.Name, res[i].Interface().(error))
	}

	return res[0]
}

func (s *state) checkFuncReturn(name string, t reflect.Type) {
	// Check the number of return values
	if t.NumOut() == 1 || (t.NumOut() == 2 && t.Out(1) == errorType) {
		return
	}

	s.errorf("can't handle multiple returns from function %s", name)
}

func (s *state) evalArg(t reflect.Type, n Node) reflect.Value {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, ok := n.(*NumberNode); ok && n.IsInt {
			v := reflect.New(t).Elem()
			v.SetInt(n.Int64)
			return v
		}
		s.errorf("expected integer; found %s", n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, ok := n.(*NumberNode); ok && n.IsUint {
			v := reflect.New(t).Elem()
			v.SetUint(n.Uint64)
			return v
		}
		s.errorf("expected unsigned integer; found %s", n)

	case reflect.Interface:
		if t.NumMethod() == 0 {
			s.evalEmptyInterface(t, n)
		}
	}

	s.errorf("can't handle %s for arg of type %s", n, t)
	panic("not reached")
}

func (s *state) evalEmptyInterface(t reflect.Type, n Node) reflect.Value {
	switch n := n.(type) {
	case *NumberNode:
		return s.idealConstant(n)
	}

	s.errorf("can't handle assignment of %s to empty interface argument", n)
	panic("not reached")
}

func (s *state) idealConstant(c *NumberNode) reflect.Value {
	switch {
	case c.IsInt:
		n := int(c.Int64)
		if int64(n) != c.Int64 {
			s.errorf("%s overflows int", c.Text)
		}
		return reflect.ValueOf(n)

	case c.IsUint:
		s.errorf("unsigned integers are not supported: %s", c.Text)
	}

	return zero
}

func (s *state) print(v reflect.Value) {
	fmt.Fprintln(s.output, v.Interface())
}

func Exec(output io.Writer, tree *ListNode, funcs map[string]interface{}) (err error) {
	// Convert the functions to reflect values
	f := map[string]reflect.Value{}
	for name, fn := range funcs {
		f[name] = reflect.ValueOf(fn)
	}

	// Build the environment
	s := &state{
		vars:   make(variables),
		funcs:  f,
		output: output,
		t:      tree,
	}

	// Hook up the recover
	defer s.recover(&err)

	// Start the execution
	s.print(s.exec())

	return nil
}

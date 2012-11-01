package main

import (
	"fmt"
	"io"
	"reflect"
	"runtime"
)

type lambdaValue struct {
	args []string
	body *CallNode
}

func (v *lambdaValue) String() string {
	return fmt.Sprintf("<lambda value with arity %d>", len(v.args))
}

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
	zero      reflect.Value

	lambdaType = reflect.TypeOf((*lambdaValue)(nil))
)

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
	for _, n := range s.t.Nodes {
		s.print(s.walkNode(n))
	}

	return nil
}

// ========================================================

type variables map[string]reflect.Value
type functions map[string]reflect.Value

type state struct {
	funcs  functions
	vars   variables
	output io.Writer
	t      *ListNode
	outer  *state
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

func (s *state) print(v reflect.Value) {
	// Don't print the zero value
	if v == zero {
		return
	}

	// Strings are printed as they come, without newlines
	if v.Kind() == reflect.String {
		fmt.Fprint(s.output, v.Interface())
		return
	}

	// The rest of values are printed with a newline
	// (in prevention of an object printing)
	fmt.Fprintln(s.output, v.Interface())
}

func (s *state) walkNode(n Node) reflect.Value {
	switch n := n.(type) {
	case *CallNode:
		return s.walkCall(n)

	case *DefineNode:
		return s.walkDefine(n)

	case *SetNode:
		return s.walkSet(n)

	case *IfNode:
		return s.walkIf(n)

	case *BeginNode:
		return s.walkBegin(n)

	case *VarNode:
		return s.walkVar(n)

	case *BoolNode:
		return s.walkBool(n)

	case *NumberNode:
		return s.walkNumber(n)

	case *StringNode:
		return s.walkString(n)

	case *LambdaNode:
		return s.walkLambda(n)
	}

	s.errorf("cannot walk the node: %s", n)
	panic("not reached")
}

func (s *state) walkCall(n *CallNode) reflect.Value {
	// Get the func in the index
	f, ok := s.evalFunction(n.Name)
	if !ok {
		f, ok = s.vars[n.Name]
		if !ok || f.Type() != lambdaType {
			s.errorf("function not defined: %s", n.Name)
		}

		// User-defined funcs are called in a different way
		return s.walkUserCall(f, n)
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
	nargs := numArgs
	if t.IsVariadic() {
		nargs += len(n.Args) - nargs
	}
	args := make([]reflect.Value, nargs)

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

	// Exec the call
	res := f.Call(args)

	// Check if the func has and returned an error
	if len(res) == 2 && !res[1].IsNil() {
		s.errorf("error calling %s: %s", n.Name, res[1].Interface().(error))
	}

	if t.NumOut() == 0 {
		return zero
	}

	return res[0]
}

func (s *state) checkFuncReturn(name string, t reflect.Type) {
	// Check the number of return values
	if t.NumOut() == 0 || t.NumOut() == 1 || (t.NumOut() == 2 && t.Out(1) == errorType) {
		return
	}

	s.errorf("can't handle multiple returns from function %s", name)
}

func (s *state) evalFunction(name string) (reflect.Value, bool) {
	for {
		if s == nil {
			break
		}

		if s.funcs != nil {
			f, ok := s.funcs[name]
			if ok {
				return f, true
			}
		}

		s = s.outer
	}
	return zero, false
}

// Return the correct value for an argument based on its needed argument.
// It gives the fixed list of types that a Go function can receive from the
// interpreter.
func (s *state) evalArg(t reflect.Type, n Node) reflect.Value {
	param := s.walkNode(n)

	// Extract the runtime types
	tparam := param.Type().Kind()
	expected := t.Kind()

	// As a special case, unsigned number should be converted from the
	// signed ones
	switch expected {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if tparam == reflect.Int {
			param = reflect.New(t).Elem()
			param.SetUint(n.(*NumberNode).Uint64)
		}
	}

	// Signal a type mismatching between the formal and current parameter
	// interface{} it's an exception, because accepts all kind of types
	if expected != reflect.Interface && expected != tparam {
		s.errorf("incorrect argument type, expected %s, got %s", expected, tparam)
	}

	return param
}

func (s *state) walkUserCall(v reflect.Value, n *CallNode) reflect.Value {
	// Check if the variable value it's really a lambda function,
	// and not any other kind of value
	if v.Type() != lambdaType {
		s.errorf("the variable %s is not a function, cannot be called", n.Name)
	}

	f := v.Interface().(*lambdaValue)

	// Check the arity of the func
	if len(f.args) != len(n.Args) {
		s.errorf("call doesn't use the correct arity: expected %d, got %s",
			len(f.args), len(n.Args))
	}

	// Create the new sub-environment
	env := &state{
		vars:   make(variables),
		output: s.output,
		outer:  s,
	}

	// Evaluate the arguments
	for i, node := range n.Args {
		env.vars[f.args[i]] = s.walkNode(node)
	}

	return env.walkCall(f.body)
}

func (s *state) walkDefine(n *DefineNode) reflect.Value {
	name := n.Variable.Name

	if _, ok := s.vars[name]; ok {
		s.errorf("variable already defined: %s", name)
	}

	s.vars[name] = s.walkNode(n.Value)
	return s.vars[name]
}

func (s *state) walkSet(n *SetNode) reflect.Value {
	name := n.Variable.Name

	if _, ok := s.vars[name]; !ok {
		s.errorf("variable not defined: %s", name)
	}

	s.vars[name] = s.walkNode(n.Value)
	return s.vars[name]
}

func (s *state) walkIf(n *IfNode) reflect.Value {
	test := s.walkNode(n.Test)
	if test.Kind() == reflect.Bool {
		if test.Bool() {
			return s.walkNode(n.Conseq)
		} else {
			return s.walkNode(n.Alt)
		}
	}

	s.errorf("if condition is not a boolean")
	panic("not reached")
}

func (s *state) walkBegin(n *BeginNode) (v reflect.Value) {
	for _, node := range n.Nodes {
		v = s.walkNode(node)
	}
	return
}

func (s *state) walkNumber(c *NumberNode) reflect.Value {
	switch {
	case c.IsInt:
		n := int(c.Int64)
		if int64(n) != c.Int64 {
			s.errorf("%s overflows int", c.Text)
		}
		return reflect.ValueOf(n)
	}

	panic("not reached")
}

func (s *state) walkVar(n *VarNode) reflect.Value {
	value, ok := s.vars[n.Name]
	if !ok {
		s.errorf("variable not defined: %s", n.Name)
	}
	return value
}

func (s *state) walkBool(n *BoolNode) reflect.Value {
	return reflect.ValueOf(n.Value)
}

func (s *state) walkString(n *StringNode) reflect.Value {
	return reflect.ValueOf(n.Text)
}

func (s *state) walkLambda(n *LambdaNode) reflect.Value {
	c := &lambdaValue{
		args: make([]string, len(n.Args)),
		body: n.Body.(*CallNode),
	}

	for i, arg := range n.Args {
		c.args[i] = arg.(*VarNode).Name
	}

	return reflect.ValueOf(c)
}

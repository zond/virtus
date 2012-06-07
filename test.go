
package main

/*
#cgo CFLAGS: -I/opt/local/include/js -DXP_UNIX
#cgo LDFLAGS: -L/opt/local/lib -ljs
#include <jsapi.h>
*/
import "C"

import (
	"fmt"
)

var runtime *C.JSRuntime
var scripts map[*JS]bool = make(map[*JS]bool)

func init() {
	runtime = C.JS_NewRuntime(1024 * 1024)
}

func Shutdown() {
	for script,_ := range(scripts) {
		script.Destroy()
	}
	C.JS_DestroyRuntime(runtime)
	C.JS_ShutDown()
}

type JS struct {
	context *C.JSContext
	global *C.JSObject
}

type JSObject struct {
	js *JS
	value *C.JSObject
}

type JSFunction struct {
	js *JS
	value *C.JSFunction
}

func NewJS() *JS {
	context := C.JS_NewContext(runtime, 8192)
	global := C.JS_NewObject(context, nil, nil, nil)
	script := &JS{context, global}
	C.JS_InitStandardClasses(context, global)
	scripts[script] = true
	return script
}

func (self *JS) Destroy() {
	delete(scripts, self)
	C.JS_DestroyContext(self.context)
}

func (self *JS) convert(val C.jsval) interface{} {
	t := C.JS_TypeOfValue(self.context, val)
	if t == C.JSTYPE_VOID {
		return nil
	} else if t == C.JSTYPE_OBJECT {
		var obj C.JSObject
		var obj_p = &obj
		C.JS_ValueToObject(self.context, val, &obj_p)
		return &JSObject{self, &obj}
	} else if t == C.JSTYPE_FUNCTION {
		return &JSFunction{self, C.JS_ValueToFunction(self.context, val)}
	} else if t == C.JSTYPE_STRING {
		return C.GoString(C.JS_GetStringBytes(C.JS_ValueToString(self.context, val)))
	} else if t == C.JSTYPE_NUMBER {
 		var rval C.jsdouble
		C.JS_ValueToNumber(self.context, val, &rval)
		return float64(rval)
	} else if t == C.JSTYPE_BOOLEAN {
		var rval C.JSBool
		C.JS_ValueToBoolean(self.context, val, &rval)
		return rval == C.JS_TRUE
	}
	return nil
}

func (self *JS) Eval(script string) interface{} {
	var rval C.jsval
	C.JS_EvaluateScript(self.context, 
                self.global,
                C.CString(script), 
                C.uintN(len(script)),
                C.CString("script"),
                1, 
                &rval)
	return self.convert(rval)
}

func main() {
	script := NewJS()

	x := script.Eval("x = 10; y = x * x; y == 100; \"martin\"; function(i) { return i + 1; }; new Object()")
	fmt.Println("hej", x)
	Shutdown()
}

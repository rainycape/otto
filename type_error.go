package otto

import "strings"

func (rt *_runtime) newErrorObject(name string, message Value) *_object {
	self := rt.newClassObject("Error")
	if message.IsDefined() {
		msg := message.string()
		self.defineProperty("message", toValue_string(msg), 0111, false)
		self.value = newError(rt, name, strings.Replace(msg, "%", "%%", -1))
	} else {
		self.value = newError(rt, name)
	}
	return self
}

package glman

import (
	"fmt"
	"tetra/lib/dbg"
	"unsafe"

	"tetra/internal/gl"
)

// DbgCheckError invoke glGetError and verify the return code, panic if got a error
func DbgCheckError() {
	x := gl.GetError()
	if x == gl.NO_ERROR {
		return
	}
	var s string
	switch x {
	case gl.INVALID_ENUM:
		s = "GL_INVALID_ENUM An unacceptable value is specified for an enumerated argument."
	case gl.INVALID_VALUE:
		s = "GL_INVALID_VALUE A numeric argument is out of range."
	case gl.INVALID_OPERATION:
		s = "GL_INVALID_OPERATION The specified operation is not allowed in the current state."
	case gl.STACK_OVERFLOW:
		s = "GL_STACK_OVERFLOW This function would cause a stack overflow."
	case gl.STACK_UNDERFLOW:
		s = "GL_STACK_UNDERFLOW This function would cause a stack underflow."
	case gl.OUT_OF_MEMORY:
		s = "GL_OUT_OF_MEMORY There is not enough memory left to execute the function."
	// case gl.TABLE_TOO_LARGE:
	// 	s := "GL_TABLE_TOO_LARGE The specified table exceeds the implementation's maximum supported table size."
	default:
		s = fmt.Sprintf("unkown opengl error %d", x)
	}
	panic(s)
}

func debugProc(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {
	var srcStr string
	switch source {
	case gl.DEBUG_SOURCE_API:
		srcStr = "API"
	case gl.DEBUG_SOURCE_WINDOW_SYSTEM:
		srcStr = "WINDOW_SYSTEM"
	case gl.DEBUG_SOURCE_SHADER_COMPILER:
		srcStr = "SHADER_COMPILER"
	case gl.DEBUG_SOURCE_THIRD_PARTY:
		srcStr = "THIRD_PARTY"
	case gl.DEBUG_SOURCE_APPLICATION:
		srcStr = "APPLICATION"
	//case gl.DEBUG_SOURCE_OTHER:
	//	srcStr = "OTHER"
	default:
		srcStr = "OTHER"
	}
	var typStr string
	switch gltype {
	case gl.DEBUG_TYPE_ERROR:
		typStr = "ERROR"
	case gl.DEBUG_TYPE_DEPRECATED_BEHAVIOR:
		typStr = "DEPRECATED"
	case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:
		typStr = "UNDEFINED"
	case gl.DEBUG_TYPE_PORTABILITY:
		typStr = "PORTABILITY"
	case gl.DEBUG_TYPE_PERFORMANCE:
		typStr = "PERFORMANCE"
	case gl.DEBUG_TYPE_MARKER:
		typStr = "MARKER"
	case gl.DEBUG_TYPE_PUSH_GROUP:
		typStr = "PUSH_GROUP"
	case gl.DEBUG_TYPE_POP_GROUP:
		typStr = "POP_GROUP"
	//case gl.DEBUG_TYPE_OTHER:
	default:
		typStr = "OTHER"
	}

	var svtyStr string
	switch severity {
	case gl.DEBUG_SEVERITY_LOW:
		svtyStr = "LOW"
	//case gl.DEBUG_SEVERITY_MEDIUM: svtyStr="";
	case gl.DEBUG_SEVERITY_HIGH:
		svtyStr = "HIGH"
	default:
		svtyStr = "MEDIUM"
	}
	dbg.Logf("%s %s %s: %d: %s\n", srcStr, typStr, svtyStr, id, message)
}

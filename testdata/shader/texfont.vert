// texture font

uniform mat4 uniMatP;
uniform mat4 uniMatV;
uniform mat4 uniMatM;

in vec3 attPos;
in vec2 attTC;

out vec2 vryTC;
out vec2 vryPos;

void main() {
  vryTC = attTC;
  gl_Position = uniMatP * uniMatV * uniMatM * vec4(attPos, 1);
  vryPos = gl_Position.xy;
}

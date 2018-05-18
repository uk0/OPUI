// texture font without edge

uniform sampler2D uniTex0; // glyph texture, alpha only
uniform float uniTexSize; // size of texture
uniform vec4 uniColors[2]; // output text color
uniform vec4 uniClip2D;  // clip rect [l,t,r,b]

in vec2 vryPos;
//in vec2 vryTC;

// 2D clip on NDC space
float rectClip(vec2 pt) {
  // NDC is y-up, our 2D is y-down, so clip[3] is top, clip[1] is bottom
  return step(uniClip2D[0], pt.x) * step(uniClip2D[3], pt.y) *
    step(pt.x, uniClip2D[2]) * step(pt.x, uniClip2D[1]);
}

void main() {
  //float bodyAlpha = texture2D(uniTex0, vryTC).w;
  gl_FragColor = uniColors[0] * rectClip(vryPos);
}

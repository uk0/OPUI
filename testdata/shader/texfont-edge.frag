// texture font with edge

uniform sampler2D uniTex0; // glyph texture, alpha only
uniform float uniTexSize; // size of texture
uniform vec4 uniColors[2]; // output text color
uniform vec4 uniClip2D;  // clip rect [l,t,r,b]

in vec2 vryPos;
in vec2 vryTC;

// 2D clip on NDC space
float rectClip(vec2 pt) {
  // NDC is y-up, our 2D is y-down, so clip[3] is top, clip[1] is bottom
  return step(uniClip2D[0], pt.x) * step(uniClip2D[3], pt.y) *
    step(pt.x, uniClip2D[2]) * step(pt.x, uniClip2D[1]);
}

// max alpha of neighbour texels
float edgeDetect(vec2 pt) {
  float inv = 1/uniTexSize;
  pt = pt * uniTexSize;
  float a = texture2D(uniTex0,vec2(pt.x, pt.y+1)*inv).w;
  float b = texture2D(uniTex0,vec2(pt.x, pt.y-1)*inv).w;
  float c = texture2D(uniTex0,vec2(pt.x+1, pt.y)*inv).w;
  float d = texture2D(uniTex0,vec2(pt.x-1, pt.y)*inv).w;
  return clamp(a + b + c + d, 0, 1);
}

void main() {
  float bodyAlpha = texture2D(uniTex0, vryTC).w;
  float edgeAlpha = edgeDetect(vryTC);
  gl_FragColor = mix(uniColors[1]*edgeAlpha, uniColors[0], bodyAlpha) * rectClip(vryPos);
}

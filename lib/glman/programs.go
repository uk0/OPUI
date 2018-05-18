package glman

var (
	//progTest        *Program
	progTexFont     *Program
	progTexFontEdge *Program
	progSimpleDraw  *Program

	//progSimpleTex   *Program
	//progColorDraw *Program
)

// UseProgTexFont load and use the tex font program
func UseProgTexFont(edge bool) (p *Program) {
	if edge {
		if progTexFontEdge == nil {
			progTexFontEdge = MustLoadProgram("texfont.vert", "texfont-edge.frag")
		}
		p = progTexFontEdge
	} else {
		if progTexFont == nil {
			progTexFont = MustLoadProgram("texfont.vert", "texfont.frag")
		}
		p = progTexFont
	}
	p.UseProgram()
	DbgCheckError()
	p.LoadMVPStack()
	p.LoadClip2DStack()
	return p
}

// UseProgSimpleDraw load and use the simple draw program
func UseProgSimpleDraw() (p *Program) {
	if progSimpleDraw == nil {
		progSimpleDraw = MustLoadProgram("simple-draw.vert", "simple-draw.frag")
	}
	p = progSimpleDraw
	p.UseProgram()
	p.LoadMVPStack()
	p.LoadClip2DStack()
	return p
}

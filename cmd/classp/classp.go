package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

var gopath = filepath.Clean(os.Getenv("GOPATH"))
var targetPkg string                   // package path to generate for
var targetImps = make(map[string]bool) // imports to generate
var verbose bool
var factory string

var classes = make(map[string]*class)  // class classes, struct
var ifaces = make(map[string]bool)     // manual defined interfaces
var pkgnames = make(map[string]string) // parsed packages

type method struct {
	doc string
	sig string
	out string
}

type class struct {
	name  string
	super []string
	//extSuper []string
	child   []string
	methods map[string]method

	self map[string]bool

	explicitClass bool
}

func (c *class) isClass() bool {
	return c.explicitClass || c.super != nil || c.child != nil
}

func isSameFunc(a, b method) bool {
	return a.sig == b.sig
}

func (c *class) reduce(s *class, all map[string]method) {
	var nkname = nakedName(c.name)
	var spname = nakedName(s.name)
	merged := make(map[string]method)
	for n, t0 := range all {
		if t1, ok := c.methods[n]; ok {
			if isSameFunc(t0, t1) {
				delete(c.methods, n)
			} else {
				fmt.Printf("error: class %s override method mismatch:\n", nkname)
				fmt.Printf("   func%s  // %s\n", t1.sig, c.name)
				fmt.Printf("   func%s  // %s\n", t0.sig, s.name)
			}
		}
		merged[n] = t0
	}

	for n, t := range c.methods {
		merged[n] = t
	}

	for self := range s.self {
		c.self[spname+"."+self] = true
	}

	for _, d := range c.child {
		classes[d].reduce(c, merged)
	}
}

func reduce() {
	for _, c := range classes {
		if c.super != nil {
			continue
		}
		for _, d := range c.child {
			classes[d].reduce(c, c.methods)
		}
	}
}

func init() {
	flag.BoolVar(&verbose, "v", false, "Print verbose message")
	flag.StringVar(&factory, "factory", "tetra/lib/factory",
		"The factory package path")

}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", name)
	fmt.Fprintf(os.Stderr, "  %s [OPTIONS] [DIR]\n\n", name)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
}

func isClassDoc(s string) bool {
	return strings.Index(s, "class") != -1
}

func isClassComment(s string) bool {
	return strings.Index(s, "class") != -1 || strings.Index(s, "super") != -1
}

func fnSigStr(fileimps map[string]string, expr ast.Expr) string {
	switch p := expr.(type) {
	// case nil:
	// 	return ""
	case *ast.StarExpr:
		return "*" + fnSigStr(fileimps, p.X)
	case *ast.SelectorExpr:
		if ident, ok := p.X.(*ast.Ident); ok {
			ppath, ok := fileimps[ident.Name]
			if ok {
				return fullName(ppath, p.Sel.Name)
			}
		}
		return fnSigStr(fileimps, p.X) + "." + p.Sel.Name
	case *ast.Ident:
		return p.Name
	case *ast.ArrayType:
		if p.Len == nil {
			return fmt.Sprintf("[]%s", fnSigStr(fileimps, p.Elt))
		}
		return fmt.Sprintf("[%s]%s", p.Len, fnSigStr(fileimps, p.Elt))
	case *ast.InterfaceType:
		// orders matter
		var list []string
		for _, m := range p.Methods.List {
			list = append(list, m.Names[0].Name+fnSigStr(fileimps, m.Type))
		}
		sort.Strings(list)
		return "interface{" + strings.Join(list, ";") + "}"
	case *ast.FuncType:
		s := "("
		for i, arg := range p.Params.List {
			if i > 0 {
				s += ","
			}
			for j := range arg.Names {
				if j > 0 {
					s += ", "
				}
				s += fnSigStr(fileimps, arg.Type)
			}
		}
		s += ")"
		if p.Results == nil {
			// no return
		} else if p.Results.NumFields() == 1 {
			ret := p.Results.List[0]
			s += fnSigStr(fileimps, ret.Type)
		} else if p.Results.NumFields() > 1 {
			s += "("
			for i, ret := range p.Results.List {
				if i > 0 {
					s += ","
				}
				for j := range ret.Names {
					if j > 0 {
						s += ", "
					}
					s += fnSigStr(fileimps, ret.Type)
				}
			}
			s += ")"
		}
		return s
	default:
		panic(fmt.Sprintf("unsupported type expr %#v", expr))
	}
}

func fnOutStr(fileimps map[string]string, expr ast.Expr) string {
	switch p := expr.(type) {
	// case nil:
	// 	return ""
	case *ast.StarExpr:
		return "*" + fnOutStr(fileimps, p.X)
	case *ast.SelectorExpr:
		if ident, ok := p.X.(*ast.Ident); ok {
			ppath, ok := fileimps[ident.Name]
			//fmt.Printf("%s => %s\n", ident.Name, ppath)
			if ok {
				return nameForTarget(fullName(ppath, p.Sel.Name))
			}
		}
		return fnOutStr(fileimps, p.X) + "." + p.Sel.Name
	case *ast.Ident:
		return p.Name
	case *ast.ArrayType:
		if p.Len == nil {
			return fmt.Sprintf("[]%s", fnOutStr(fileimps, p.Elt))
		}
		return fmt.Sprintf("[%s]%s", p.Len, fnOutStr(fileimps, p.Elt))
	case *ast.InterfaceType:
		// orders matter
		var list []string
		for _, m := range p.Methods.List {
			list = append(list, m.Names[0].Name+fnOutStr(fileimps, m.Type))
		}
		sort.Strings(list)
		return "interface{" + strings.Join(list, ";") + "}"
	case *ast.FuncType:
		s := "("
		for i, arg := range p.Params.List {
			if i > 0 {
				s += ", "
			}
			for j, n := range arg.Names {
				if j > 0 {
					s += ", "
				}
				s += n.Name
			}
			if len(arg.Names) > 0 {
				s += " "
			}
			s += fnOutStr(fileimps, arg.Type)
		}
		s += ")"
		//fmt.Printf("%#v\n", p)
		if p.Results == nil {
			// no return
		} else if p.Results.NumFields() == 1 {
			ret := p.Results.List[0]
			s += " " + fnOutStr(fileimps, ret.Type)
		} else if p.Results.NumFields() > 1 {
			s += " ("
			for i, ret := range p.Results.List {
				if i > 0 {
					s += ", "
				}
				for j, n := range ret.Names {
					if j > 0 {
						s += ", "
					}
					s += n.Name
				}
				if len(ret.Names) > 0 {
					s += " "
				}
				s += fnOutStr(fileimps, ret.Type)
			}
			s += ")"
		}
		return s
	default:
		panic(fmt.Sprintf("unsupported type expr %#v", expr))
	}
}

func dirToPkg(dir string) string {
	s, err := filepath.Rel(gopath+"/src", dir)
	if err != nil {
		panic(err)
	}
	return strings.Replace(s, "\\", "/", -1)
}

func pkgToDir(pkg string) string {
	return filepath.Clean(gopath + "/src/" + pkg)
}

func isPkgInGOPATH(pkg string) bool {
	info, err := os.Stat(pkgToDir(pkg))
	if err == nil {
		return info.IsDir()
	}
	return false
}

func isInTargetPkg(typeName string) bool {
	return strings.HasPrefix(typeName, targetPkg)
}

func pkgPath(typeName string) string {
	pos := strings.LastIndexByte(typeName, '.')
	if pos == -1 {
		return typeName
	}
	return typeName[:pos]
}

func nakedName(typeName string) string {
	pos := strings.LastIndexByte(typeName, '.')
	if pos == -1 {
		return typeName
	}
	return typeName[pos+1:]
}

func fullName(pkgpath string, nakedName string) string {
	return strings.Replace(pkgpath, "\\", "/", -1) + "." + nakedName
}

func nameForTarget(typeName string) string {
	if isInTargetPkg(typeName) {
		return nakedName(typeName)
	}
	s := pkgPath(typeName)
	pkg, ok := pkgnames[s]
	if !ok {
		pkg = filepath.Base(s)
	}
	return pkg + "." + nakedName(typeName)
}

func toIname(name string) string {
	pos := strings.LastIndexByte(name, '.')
	if pos == -1 {
		return "I" + name
	}
	return name[:pos+1] + "I" + name[pos+1:]
}

func loadPkg(dir string) (pkgname string) {
	pkgpath := dirToPkg(dir)
	if verbose {
		log.Println("parse", pkgpath)
	}
	fset := new(token.FileSet)
	pm, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalln(err)
	}

	if len(pm) == 0 {
		log.Fatalln("no go source in dir.")
	}
	if len(pm) > 1 {
		log.Fatalln("mutipile package in same dir.")
	}

	var pkg *ast.Package
	for _, pkg = range pm {
		break // only first package
	}
	pkgname = strings.Replace(pkg.Name, "\\", "/", -1)

	// imports

	// find out structs and interfaces
	for _, file := range pkg.Files {
		fileimps := make(map[string]string) // "lang" => "tetra/lib/lang"
		for _, imp := range file.Imports {
			s, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				log.Fatalln(err)
			}
			var n string
			if imp.Name == nil {
				n = filepath.Base(s)
			} else {
				n = imp.Name.Name
			}
			fileimps[n] = s
			//fmt.Printf("%s => %s\n", n, s)
		}

		ast.Inspect(file, func(node ast.Node) bool {
			var decl *ast.GenDecl
			var ok bool
			if decl, ok = node.(*ast.GenDecl); !ok {
				return true
			}
			if decl.Tok != token.TYPE {
				return true
			}

			declHasClassDoc := decl.Doc != nil && isClassDoc(decl.Doc.Text())
			// type X struct {}
			// type (
			//   Foo struct {}
			//   Bar struct {}
			//   IFoo interface{}
			// )
			for _, spec := range decl.Specs {
				typ := spec.(*ast.TypeSpec)
				if st, ok := typ.Type.(*ast.StructType); ok {
					if st.Incomplete {
						continue
					}
					name := fullName(pkgpath, typ.Name.Name)
					c := new(class)
					c.self = make(map[string]bool)
					c.methods = make(map[string]method)
					c.methods["Class"] = method{
						doc: "Class name for factory",
						sig: "()string",
						out: "() string",
					}
					c.name = name
					c.explicitClass = declHasClassDoc
					for _, field := range st.Fields.List {
						if field.Names != nil {
							// type MyWindow struct {
							//   Self interface{}
							// }
							if func(expr ast.Expr) bool {
								switch x := expr.(type) {
								case *ast.InterfaceType:
									return true
								case *ast.Ident:
									return strings.HasPrefix(x.Name, "I")
								case *ast.SelectorExpr:
									return strings.HasPrefix(x.Sel.Name, "I")
								default:
									return false
								}
							}(field.Type) {
								for _, name := range field.Names {
									if name.Name == "Self" {
										c.self["Self"] = true
										break
									}
								}
							}
						} else if sel, ok := field.Type.(*ast.SelectorExpr); ok {
							if field.Comment != nil && isClassComment(field.Comment.Text()) {
								// type MyWindow struct {
								//   winl.Window // super
								// }
								pnx, ok := sel.X.(*ast.Ident)
								if !ok {
									continue
								}
								ppath, ok := fileimps[pnx.Name]
								if !ok {
									continue
								}
								if _, ok = pkgnames[ppath]; !ok {
									if !isPkgInGOPATH(ppath) {
										panic("pkg not in GOPATH: " + ppath)
									}
									pkgnames[ppath] = "-"
									pkgnames[ppath] = loadPkg(pkgToDir(ppath))
								}
								c.super = append(c.super, fullName(ppath, sel.Sel.Name))
							}
						} else if ident, ok := field.Type.(*ast.Ident); ok {
							// type MyWindow struct {
							//   Window
							// }
							c.super = append(c.super, fullName(pkgpath, ident.Name))
						}
					}
					classes[name] = c
				} else if iface, ok := typ.Type.(*ast.InterfaceType); ok {
					if iface.Incomplete {
						continue
					}
					name := fullName(pkgpath, typ.Name.Name)
					ifaces[name] = true
				}
			}
			return true
		})
	}

	// find out methods
	for _, file := range pkg.Files {
		fileimps := make(map[string]string) // "lang" => "tetra/lib/lang"
		for _, imp := range file.Imports {
			s, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				log.Fatalln(err)
			}
			var n string
			if imp.Name == nil {
				n = filepath.Base(s)
			} else {
				n = imp.Name.Name
			}
			fileimps[n] = s
			//fmt.Printf("%s => %s\n", n, s)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if fdecl, ok := node.(*ast.FuncDecl); ok {
				if fdecl.Recv == nil || fdecl.Recv.NumFields() != 1 {
					return true
				}
				// recv must be pointer
				star, ok := fdecl.Recv.List[0].Type.(*ast.StarExpr)
				if !ok {
					return true
				}
				// recv must be local type *Foo, not *bar.Foo
				recvIdent, ok := star.X.(*ast.Ident)
				if !ok {
					return true
				}
				c, ok := classes[fullName(pkgpath, recvIdent.Name)]
				if !ok {
					// not a class type
					return true
				}
				if !unicode.IsUpper(rune(fdecl.Name.Name[0])) {
					// private function
					return true
				}
				var m method
				if fdecl.Doc != nil {
					m.doc = strings.TrimSpace(fdecl.Doc.Text())
				}
				m.sig = fnSigStr(fileimps, fdecl.Type)
				m.out = fnOutStr(fileimps, fdecl.Type)
				// if fdecl.Name.Name == "Size" {
				// 	fmt.Printf("%s\n", m.sig)
				// 	fmt.Printf("%s\n", m.out)
				// 	fmt.Printf("%#v\n", fdecl.Type.Results.List[0].Names[1].String())
				// }
				c.methods[fdecl.Name.Name] = m
			}
			return true
		})
	}
	return
}

func fillChild() {
	for n, c := range classes {
		for _, s := range c.super {
			sc := classes[s]
			if sc != nil {
				sc.child = append(sc.child, n)
			}
		}
	}
}

func fixImports(filename string) {
	_, err := exec.LookPath("goimports")
	if err != nil {
		cmd1 := exec.Command("go", "get", "golang.org/x/tools/cmd/goimports")
		if verbose {
			log.Println("go", "get", "golang.org/x/tools/cmd/goimports")
		}
		if buf, err1 := cmd1.CombinedOutput(); err1 != nil {
			log.Println(buf)
			log.Fatalln(err1)
		}
	}

	cmd := exec.Command("goimports", "-w", filename)
	if verbose {
		log.Println("goimports", "-w", filename)
	}
	if buf, err := cmd.CombinedOutput(); err != nil {
		log.Println(buf)
		log.Fatalln(err)
	}
}

func main() {
	if gopath == "" {
		panic("unkown GOPATH")
	}
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() > 1 {
		usage()
		os.Exit(2)
	}
	if flag.NArg() == 1 {
		if err := os.Chdir(flag.Arg(0)); err != nil {
			log.Fatalln(err)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	targetPkg = dirToPkg(dir)
	dirBaseName := filepath.Base(dir)
	filename := "z-class-" + dirBaseName + ".go"
	os.Remove(dir + "/" + filename)

	pkgname := loadPkg(dir)
	// if pkgname == "main" && regexp.MustCompile("[a-z]+").MatchString(dirBaseName) {
	// 	pkgname = dirBaseName
	// }

	fillChild()
	reduce()

	var names []string
	for name := range classes {
		names = append(names, name)
	}
	sort.Strings(names)

	defer fixImports(filename)

	if verbose {
		log.Println("write", filename)
	}
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	fmt.Fprintln(file, `package `+pkgname)
	fmt.Fprintf(file, "\n")
	fmt.Fprintln(file, `// Auto generated file, do NOT edit!`)
	fmt.Fprintf(file, "\n")
	fmt.Fprintln(file, `import "`+factory+`"`)

	fmt.Fprintf(file, "\n")
	fmt.Fprintln(file, `var factoryRegisted bool`)
	fmt.Fprintf(file, "\n")
	fmt.Fprintln(file, `// FactoryRegister register creator in factory for package `+pkgname)
	fmt.Fprintln(file, `func FactoryRegister() {`)
	fmt.Fprintln(file, "\tif factoryRegisted {")
	fmt.Fprintln(file, "\t\treturn")
	fmt.Fprintln(file, "\t}")
	fmt.Fprintln(file, "\tfactoryRegisted = true")
	fmt.Fprintf(file, "\n")
	for _, name := range names {
		c := classes[name]
		if !c.isClass() {
			continue
		}
		if !isInTargetPkg(name) {
			continue
		}
		tname := nameForTarget(c.name)
		fmt.Fprintf(file, "\tfactory.Register(`%s.%s`, func() interface{} {\n", pkgname, tname)
		fmt.Fprintf(file, "\t\treturn New%s()\n", tname)
		fmt.Fprintf(file, "\t})\n")
	}
	fmt.Fprintf(file, "}\n")

	for _, name := range names {
		c := classes[name]
		if !c.isClass() {
			continue
		}
		if !isInTargetPkg(name) {
			continue
		}
		tname := nameForTarget(c.name)
		// NewFoo function
		fmt.Fprintf(file, "\n")
		fmt.Fprintf(file, "// New%s create and init new %s object.\n", tname, tname)
		fmt.Fprintf(file, "func New%s() *%s {\n", tname, tname)
		fmt.Fprintf(file, "\tp := new(%s)\n", tname)
		for self := range c.self {
			fmt.Fprintf(file, "\tp.%s = p\n", self)
		}

		fmt.Fprintf(file, "\tp.Init()\n")
		fmt.Fprintf(file, "\treturn p\n")
		fmt.Fprintf(file, "}\n")

		//
		fmt.Fprintf(file, "\n")
		fmt.Fprintf(file, "// Class name for factory\n")
		fmt.Fprintf(file, "func (p *%s) Class() string {\n", tname)
		fmt.Fprintf(file, "\treturn (`%s.%s`)\n", pkgname, tname)
		fmt.Fprintf(file, "}\n")

		// Interface
		iname := toIname(tname)
		if _, ok := ifaces[iname]; ok {
			fmt.Fprintf(file, "\n")
			fmt.Fprintf(file, "// %s already defined, no auto generate code.\n", iname)
		} else {
			fmt.Fprintf(file, "\n")
			fmt.Fprintf(file, "// %s is interface of class %s\n", iname, tname)
			fmt.Fprintf(file, "type %s interface {\n", iname)
			// for _, s := range c.extSuper {
			// 	fmt.Fprintf(file, "\t%s\n", s)
			// }
			for _, s := range c.super {
				fmt.Fprintf(file, "\t%s\n", toIname(nameForTarget(s)))
			}
			var ns []string
			for n := range c.methods {
				ns = append(ns, n)
			}
			sort.Strings(ns)
			for _, n := range ns {
				m := c.methods[n]
				if m.doc != "" {
					fmt.Fprintf(file, "\t// %s\n", m.doc)
				}
				fmt.Fprintf(file, "\t%s%s\n", n, m.out)
			}
			fmt.Fprintf(file, "}\n")
		}
	}
}

package renderer

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// KindMermaidBlock is the AST node kind for Mermaid diagram blocks.
var KindMermaidBlock = ast.NewNodeKind("MermaidBlock")

// MermaidBlock is an AST node representing a Mermaid diagram.
type MermaidBlock struct {
	ast.BaseBlock
	Code []byte
}

// Kind implements ast.Node.
func (b *MermaidBlock) Kind() ast.NodeKind { return KindMermaidBlock }

// Dump implements ast.Node.
func (b *MermaidBlock) Dump(source []byte, level int) {
	ast.DumpHelper(b, source, level, nil, nil)
}

// mermaidASTTransformer replaces fenced code blocks with language "mermaid"
// with MermaidBlock nodes so they are rendered as <div class="mermaid">.
type mermaidASTTransformer struct{}

func (t *mermaidASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	var toReplace []ast.Node
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindFencedCodeBlock {
			return ast.WalkContinue, nil
		}
		fcb := n.(*ast.FencedCodeBlock)
		if string(fcb.Language(source)) == "mermaid" {
			toReplace = append(toReplace, n)
		}
		return ast.WalkContinue, nil
	})

	for _, n := range toReplace {
		fcb := n.(*ast.FencedCodeBlock)
		var buf bytes.Buffer
		lines := fcb.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			buf.Write(line.Value(source))
		}
		mb := &MermaidBlock{Code: buf.Bytes()}
		parent := n.Parent()
		parent.InsertBefore(parent, n, mb)
		parent.RemoveChild(parent, n)
	}
}

// mermaidBlockRenderer renders MermaidBlock nodes as <div class="mermaid"> elements.
type mermaidBlockRenderer struct{}

func (r *mermaidBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMermaidBlock, r.render)
}

func (r *mermaidBlockRenderer) render(
	w util.BufWriter, _ []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	mb := node.(*MermaidBlock)
	_, _ = w.WriteString(`<div class="mermaid">`)
	_, _ = w.Write(mb.Code)
	_, _ = w.WriteString("</div>\n")
	return ast.WalkContinue, nil
}

// MermaidExtension is a goldmark extension that converts ```mermaid code fences
// to <div class="mermaid"> elements for client-side rendering by mermaid.js.
type MermaidExtension struct{}

// Extend implements goldmark.Extender.
func (e *MermaidExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&mermaidASTTransformer{}, 50),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&mermaidBlockRenderer{}, 50),
		),
	)
}

package sempass

import (
	"shanhu.io/smlvm/dagvis"
	"shanhu.io/smlvm/lexing"
	"shanhu.io/smlvm/pl/ast"
	"shanhu.io/smlvm/pl/tast"
	"shanhu.io/smlvm/pl/types"
	"shanhu.io/smlvm/syms"
)

type builder struct {
	*lexing.ErrorList
	path string

	scope *syms.Scope

	exprFunc  func(b *builder, expr ast.Expr) tast.Expr
	constFunc func(b *builder, expr ast.Expr) tast.Expr
	stmtFunc  func(b *builder, stmt ast.Stmt) tast.Stmt
	typeFunc  func(b *builder, expr ast.Expr) types.T

	// file level dependency, for checking circular dependencies.
	deps deps

	nloop    int
	this     *tast.Ref
	thisType *types.Pointer

	retType  []types.T
	retNamed bool

	// if the parsing is in left hand side.
	// when in left hand side, referencing a variable does not count.
	lhs bool
}

func newBuilder(path string, scope *syms.Scope) *builder {
	return &builder{
		ErrorList: lexing.NewErrorList(),
		path:      path,
		scope:     scope,
	}
}

func (b *builder) lhsSwap(lhs bool) bool {
	lhs, b.lhs = b.lhs, lhs
	return lhs
}

func (b *builder) lhsRestore(lhs bool) { b.lhs = lhs }

func (b *builder) buildExpr(expr ast.Expr) tast.Expr {
	return b.exprFunc(b, expr)
}

func (b *builder) buildConstExpr(expr ast.Expr) tast.Expr {
	return b.constFunc(b, expr)
}

func (b *builder) buildConst(expr ast.Expr) *tast.Const {
	c, ok := b.buildConstExpr(expr).(*tast.Const)
	if !ok {
		b.Errorf(ast.ExprPos(expr), "expect a const")
		return nil
	}
	return c
}

func (b *builder) buildType(expr ast.Expr) types.T {
	return b.typeFunc(b, expr)
}

func (b *builder) buildStmt(stmt ast.Stmt) tast.Stmt {
	return b.stmtFunc(b, stmt)
}

func (b *builder) refSym(sym *syms.Symbol, pos *lexing.Pos) {
	if !b.lhs {
		sym.Used = true
	}

	// track file dependencies inside a package
	if b.deps == nil {
		return // no need to track deps
	}

	symPos := sym.Pos
	if symPos == nil {
		return // builtin
	}
	if sym.Pkg() != b.path {
		return // cross package reference
	}
	if pos.File == symPos.File {
		return
	}

	b.deps.add(pos.File, symPos.File)
}

func (b *builder) initDeps(asts map[string]*ast.File) {
	b.deps = newDeps(asts)
}

func (b *builder) depGraph() *dagvis.Graph { return b.deps.graph() }

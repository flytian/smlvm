package pl

import "shanhu.io/smlvm/pl/tast"

func buildSwitchStmt(b *builder, stmt *tast.SwitchStmt) {
	b.CodeErrorf(nil, "pl.notYetSupport", "switch is not supported yet")
	stmt.Expr
}

func buildFallthroughStmt(b *builder) {
}

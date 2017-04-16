package sempass

import (
	"fmt"

	"shanhu.io/smlvm/lexing"
	"shanhu.io/smlvm/pl/types"
)

type canAssignResult struct {
	err      bool
	needCast bool
	castMask []bool
}

func (r *canAssignResult) add(ok, needCast bool) {
	r.err = r.err || !ok
	r.needCast = r.needCast || needCast
	r.castMask = append(r.castMask, needCast)
}

func canAssignType(
	b *builder, pos *lexing.Pos, t types.T, right []types.T,
) *canAssignResult {
	var ts []types.T
	for range right {
		ts = append(ts, t)
	}
	return canAssigns(b, pos, ts, right)
}

func canAssigns(
	b *builder, pos *lexing.Pos, left, right []types.T,
) *canAssignResult {
	if len(left) != len(right) {
		panic("length mismatch")
	}

	res := new(canAssignResult)
	for i, t := range right {
		res.add(canAssign(b, pos, left[i], t))
	}
	return res
}

func canAssign(
	b *builder, p *lexing.Pos, left, right types.T,
) (ok bool, needCast bool) {
	if i, ok := left.(*types.Interface); ok {
		// TODO(yumuzi234): assing interface from interface
		if _, ok = right.(*types.Interface); ok {
			b.CodeErrorf(p, "pl.notYetSupported",
				"assign interface by interface is not supported yet")
			return false, false
		}
		if !assignInterface(b, p, i, right) {
			return false, false
		}
		return true, true
	}
	ok, needCast = types.CanAssign(left, right)
	if !ok {
		b.CodeErrorf(p, "pl.cannotAssign.typeMismatch",
			"cannot use %s as %s", left, right)
		return false, false
	}
	return ok, needCast
}

func assignInterface(
	b *builder, p *lexing.Pos, i *types.Interface, right types.T,
) bool {
	flag := true
	s, ok := types.PointerOf(right).(*types.Struct)
	if !ok {
		b.CodeErrorf(p, "pl.cannotAssign.interface",
			"cannot use %s as interface %s, not a struct pointer", right, i)
		return false
	}
	errorf := func(f string, a ...interface{}) {
		m := fmt.Sprintf(f, a...)
		b.CodeErrorf(p, "pl.cannotAssign.interface",
			"cannot use %s as interface %s, %s", right, i, m)
		flag = false
	}

	funcs := i.Syms.List()
	for _, f := range funcs {
		sym := s.Syms.Query(f.Name())
		if sym == nil {
			errorf("function %s not implemented", f.Name())
			continue
		}
		t2, ok := sym.ObjType.(*types.Func)
		if !ok {
			errorf("%s is a struct member but not a method", f.Name())
			continue
		}
		t2 = t2.MethodFunc
		t1 := f.ObjType.(*types.Func)
		if !types.SameType(t1, t2) {
			errorf("func signature mismatch %q, %q", t1, t2)
		}
	}
	return flag
}

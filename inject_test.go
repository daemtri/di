package di

import (
	"testing"
)

func Test_Inject(t *testing.T) {
	type testOption struct {
		addr string `flag:"addr"`
	}
	type tester struct {
		opt *testOption
	}
	newTester := func(ctx Context, opt *testOption) (*tester, error) {
		return &tester{opt: opt}, nil
	}
	ib := Inject[*tester, *testOption](newTester)
	x, err := ib.Build(nil)
	if err != nil {
		t.Errorf("build出错: %s", err)
		return
	}
	if x.opt == nil {
		t.Errorf("x.opt不能为nil")
	}
}

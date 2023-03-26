package di

import (
	"context"
	"testing"
)

func Test_Inject(t *testing.T) {
	type testOption struct {
		addr string `flag:"addr"`
	}
	type tester struct {
		opt *testOption
	}
	newTester := func(ctx context.Context, opt *testOption) (*tester, error) {
		return &tester{opt: opt}, nil
	}
	ib := Inject(newTester)
	x, err := ib.Build(nil)
	if err != nil {
		t.Errorf("build出错: %s", err)
		return
	}
	if x.opt == nil {
		t.Errorf("x.opt不能为nil")
	}
}

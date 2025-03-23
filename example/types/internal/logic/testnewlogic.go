package logic

import (
	"context"

	"github.com/linabellbiu/goctl-validate/example/types/internal/svc"
	"github.com/linabellbiu/goctl-validate/example/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestNewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestNewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestNewLogic {
	return &TestNewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestNewLogic) TestNew(req *types.TestNewReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line

	return
}

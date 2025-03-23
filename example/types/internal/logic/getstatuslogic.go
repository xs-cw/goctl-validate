package logic

import (
	"context"

	"github.com/linabellbiu/goctl-validate/example/types/internal/svc"
	"github.com/linabellbiu/goctl-validate/example/types/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetStatusLogic {
	return &GetStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetStatusLogic) GetStatus(req *types.StatusReq) (resp *types.CommonResp, err error) {
	// todo: add your logic here and delete this line

	return
}

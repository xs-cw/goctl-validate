package logic

import (
	"context"

	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/svc"
	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestTranslationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 测试翻译功能
func NewTestTranslationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestTranslationLogic {
	return &TestTranslationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestTranslationLogic) TestTranslation(req *types.TestTranslationReq) (resp *types.TestTranslationResp, err error) {
	// todo: add your logic here and delete this line

	return
}

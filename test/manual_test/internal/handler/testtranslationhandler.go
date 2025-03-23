package handler

import (
	"net/http"

	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/logic"
	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/svc"
	"github.com/linabellbiu/goctl-validate/test/manual_test/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 测试翻译功能
func TestTranslationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TestTranslationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewTestTranslationLogic(r.Context(), svcCtx)
		resp, err := l.TestTranslation(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

package handler

import (
	"net/http"

	"github.com/linabellbiu/goctl-validate/example/types/internal/logic"
	"github.com/linabellbiu/goctl-validate/example/types/internal/svc"
	"github.com/linabellbiu/goctl-validate/example/types/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateItemHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateItemReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewCreateItemLogic(r.Context(), svcCtx)
		resp, err := l.CreateItem(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

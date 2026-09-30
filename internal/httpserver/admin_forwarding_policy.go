package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"icloud-api/internal/store"
)

type forwardingPolicyRequest struct {
	VerifyForwardingTarget *bool `json:"verify_forwarding_target"`
}

func (s *Server) adminAPISetForwardingPolicy(c *gin.Context) {
	id, ok := adminAPIParseID(c)
	if !ok {
		return
	}
	var input forwardingPolicyRequest
	if !decodeAdminAPIJSON(c, &input) {
		return
	}
	if input.VerifyForwardingTarget == nil {
		writeAdminAPIError(c, http.StatusBadRequest, "VALIDATION_FAILED", "请明确指定是否校验默认转发目标")
		return
	}
	var accountErr error
	err := s.withAccountLock(c.Request.Context(), id, func() error {
		accountErr = s.store.SetAccountForwardingPolicy(c.Request.Context(), id, *input.VerifyForwardingTarget)
		return accountErr
	})
	if errors.Is(err, store.ErrNotFound) {
		writeAdminAPIError(c, http.StatusNotFound, "NOT_FOUND", "主号不存在")
		return
	}
	if err != nil {
		s.writeAdminAPIInternalError(c, err)
		return
	}
	account, err := s.store.GetAccount(c.Request.Context(), id)
	if err != nil {
		s.writeAdminAPIStoreReadError(c, err)
		return
	}
	session := mustSession(c)
	s.audit(c, &session.AdminID, session.Username, "update", "account_forwarding_policy", strconv.FormatInt(id, 10), "success", "")
	writeAdminAPIData(c, http.StatusOK, adminAPIAccountFromDomain(account))
}

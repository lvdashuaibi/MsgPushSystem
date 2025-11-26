package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lvdashuaibi/MsgPushSystem/src/constant"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/ai"
	log "github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
)

// AIPolishHandler AI内容润色处理器
type AIPolishHandler struct {
	polisher *ai.ContentPolisher
}

// NewAIPolishHandler 创建AI润色处理器
func NewAIPolishHandler(polisher *ai.ContentPolisher) *AIPolishHandler {
	return &AIPolishHandler{
		polisher: polisher,
	}
}

// PolishRequest 润色请求
type PolishRequest struct {
	OriginalIntent string `json:"original_intent" binding:"required"` // 原始意图
	Channel        int    `json:"channel"`                            // 渠道类型 (0:全部, 1:邮件, 2:短信, 3:飞书)
}

// PolishResponse 润色响应
type PolishResponse struct {
	Code    int                     `json:"code"`
	Msg     string                  `json:"msg"`
	Data    *ai.MultiChannelContent `json:"data,omitempty"`
	Content *ai.PolishedContent     `json:"content,omitempty"`
}

// OptimizeRequest 优化请求
type OptimizeRequest struct {
	Content      string `json:"content" binding:"required"` // 原始内容
	Channel      int    `json:"channel" binding:"required"` // 渠道类型
	Requirements string `json:"requirements"`               // 优化要求
}

// PolishForAllChannels 为所有渠道润色内容
// @Summary 为所有渠道润色内容
// @Description 根据原始意图，AI自动生成适配Email、SMS、飞书三种渠道的润色内容
// @Tags AI润色
// @Accept json
// @Produce json
// @Param request body PolishRequest true "润色请求"
// @Success 200 {object} PolishResponse
// @Router /ai/polish/all [post]
func (h *AIPolishHandler) PolishForAllChannels(c *gin.Context) {
	var req PolishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_SHOULD_BIND,
			Msg:  constant.GetErrMsg(constant.ERR_SHOULD_BIND),
		})
		return
	}

	log.Infof("🎨 收到多渠道润色请求，原始意图: %s", req.OriginalIntent)

	// 检查润色器是否可用
	if !h.polisher.IsAvailable() {
		log.Error("AI润色器不可用")
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "AI服务暂时不可用，请稍后重试",
		})
		return
	}

	// 执行润色
	result, err := h.polisher.PolishForAllChannels(c.Request.Context(), req.OriginalIntent)
	if err != nil {
		log.Errorf("多渠道润色失败: %v", err)
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "内容润色失败: " + err.Error(),
		})
		return
	}

	log.Infof("✅ 多渠道润色成功")
	c.JSON(http.StatusOK, PolishResponse{
		Code: constant.SUCCESS,
		Msg:  constant.GetErrMsg(constant.SUCCESS),
		Data: result,
	})
}

// PolishForSingleChannel 为单个渠道润色内容
// @Summary 为单个渠道润色内容
// @Description 根据原始意图，AI自动生成指定渠道的润色内容
// @Tags AI润色
// @Accept json
// @Produce json
// @Param request body PolishRequest true "润色请求"
// @Success 200 {object} PolishResponse
// @Router /ai/polish/single [post]
func (h *AIPolishHandler) PolishForSingleChannel(c *gin.Context) {
	var req PolishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_SHOULD_BIND,
			Msg:  constant.GetErrMsg(constant.ERR_SHOULD_BIND),
		})
		return
	}

	if req.Channel < 1 || req.Channel > 3 {
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INPUT_INVALID,
			Msg:  "渠道类型无效，必须是1(邮件)、2(短信)或3(飞书)",
		})
		return
	}

	log.Infof("🎨 收到单渠道润色请求，渠道: %d，原始意图: %s", req.Channel, req.OriginalIntent)

	// 检查润色器是否可用
	if !h.polisher.IsAvailable() {
		log.Error("AI润色器不可用")
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "AI服务暂时不可用，请稍后重试",
		})
		return
	}

	var content *ai.PolishedContent
	var err error

	// 根据渠道类型执行润色
	switch ai.ChannelType(req.Channel) {
	case ai.ChannelEmail:
		content, err = h.polisher.PolishForEmail(c.Request.Context(), req.OriginalIntent)
	case ai.ChannelSMS:
		content, err = h.polisher.PolishForSMS(c.Request.Context(), req.OriginalIntent)
	case ai.ChannelLark:
		content, err = h.polisher.PolishForLark(c.Request.Context(), req.OriginalIntent)
	default:
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INPUT_INVALID,
			Msg:  "不支持的渠道类型",
		})
		return
	}

	if err != nil {
		log.Errorf("单渠道润色失败: %v", err)
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "内容润色失败: " + err.Error(),
		})
		return
	}

	log.Infof("✅ 单渠道润色成功")
	c.JSON(http.StatusOK, PolishResponse{
		Code:    constant.SUCCESS,
		Msg:     constant.GetErrMsg(constant.SUCCESS),
		Content: content,
	})
}

// OptimizeContent 优化已有内容
// @Summary 优化已有内容
// @Description 对已有的内容进行AI优化
// @Tags AI润色
// @Accept json
// @Produce json
// @Param request body OptimizeRequest true "优化请求"
// @Success 200 {object} PolishResponse
// @Router /ai/polish/optimize [post]
func (h *AIPolishHandler) OptimizeContent(c *gin.Context) {
	var req OptimizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("参数绑定失败: %v", err)
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_SHOULD_BIND,
			Msg:  constant.GetErrMsg(constant.ERR_SHOULD_BIND),
		})
		return
	}

	if req.Channel < 1 || req.Channel > 3 {
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INPUT_INVALID,
			Msg:  "渠道类型无效，必须是1(邮件)、2(短信)或3(飞书)",
		})
		return
	}

	log.Infof("✨ 收到内容优化请求，渠道: %d", req.Channel)

	// 检查润色器是否可用
	if !h.polisher.IsAvailable() {
		log.Error("AI润色器不可用")
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "AI服务暂时不可用，请稍后重试",
		})
		return
	}

	// 执行优化
	content, err := h.polisher.OptimizeContent(
		c.Request.Context(),
		req.Content,
		ai.ChannelType(req.Channel),
		req.Requirements,
	)

	if err != nil {
		log.Errorf("内容优化失败: %v", err)
		c.JSON(http.StatusOK, PolishResponse{
			Code: constant.ERR_INTERNAL,
			Msg:  "内容优化失败: " + err.Error(),
		})
		return
	}

	log.Infof("✅ 内容优化成功")
	c.JSON(http.StatusOK, PolishResponse{
		Code:    constant.SUCCESS,
		Msg:     constant.GetErrMsg(constant.SUCCESS),
		Content: content,
	})
}

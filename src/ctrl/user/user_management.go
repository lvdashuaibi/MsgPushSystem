package user

import (
	"net/http"

	"github.com/lvdashuaibi/MsgPushSystem/src/constant"
	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/ctrlmodel"
	"github.com/lvdashuaibi/MsgPushSystem/src/ctrl/handler"
	"github.com/lvdashuaibi/MsgPushSystem/src/data"
	"github.com/lvdashuaibi/MsgPushSystem/src/pkg/log"
	"github.com/gin-gonic/gin"
)

// CreateUserHandler 创建用户处理器
type CreateUserHandler struct {
	Req  ctrlmodel.CreateUserReq
	Resp ctrlmodel.CreateUserResp
}

// CreateUser 创建用户API
func CreateUser(c *gin.Context) {
	var hd CreateUserHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	// 解析请求参数
	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("CreateUser shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	// 执行处理函数
	if err := handler.Run(&hd); err != nil {
		log.Errorf("CreateUser handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

// HandleInput 参数检查
func (h *CreateUserHandler) HandleInput() error {
	// 基本参数验证已通过binding完成
	return nil
}

// HandleProcess 处理函数
func (h *CreateUserHandler) HandleProcess() error {
	dt := data.GetData()

	// 检查用户ID是否已存在
	existingUser, err := data.UserNamespace.FindByUserID(dt.GetDB(), h.Req.UserID)
	if err == nil && existingUser != nil {
		h.Resp.Code = constant.ERR_USER_ALREADY_EXISTS
		return nil
	}

	// 创建用户对象
	user := &data.User{
		UserID:   h.Req.UserID,
		Name:     h.Req.Name,
		Nickname: h.Req.Nickname,
		Mobile:   h.Req.Mobile,
		Email:    h.Req.Email,
		LarkID:   h.Req.LarkID,
		Tags:     data.UserTags(h.Req.Tags),
		Status:   int(data.UserStatusEnabled),
	}

	// 创建用户
	if err := data.UserNamespace.Create(dt.GetDB(), user); err != nil {
		log.Errorf("创建用户失败: %s", err.Error())
		h.Resp.Code = constant.ERR_INSERT
		return err
	}

	h.Resp.UserID = h.Req.UserID
	log.Infof("用户创建成功: %s", h.Req.UserID)
	return nil
}

// GetUserHandler 获取用户处理器
type GetUserHandler struct {
	Req  ctrlmodel.GetUserReq
	Resp ctrlmodel.GetUserResp
}

// GetUser 获取用户API
func GetUser(c *gin.Context) {
	var hd GetUserHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("GetUser shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("GetUser handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *GetUserHandler) HandleInput() error {
	return nil
}

func (h *GetUserHandler) HandleProcess() error {
	dt := data.GetData()

	user, err := data.UserNamespace.FindByUserID(dt.GetDB(), h.Req.UserID)
	if err != nil {
		log.Errorf("查找用户失败: %s", err.Error())
		h.Resp.Code = constant.ERR_USER_NOT_FOUND
		return err
	}

	h.Resp.User = user
	return nil
}

// UpdateUserHandler 更新用户处理器
type UpdateUserHandler struct {
	Req  ctrlmodel.UpdateUserReq
	Resp ctrlmodel.UpdateUserResp
}

// UpdateUser 更新用户API
func UpdateUser(c *gin.Context) {
	var hd UpdateUserHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("UpdateUser shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("UpdateUser handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *UpdateUserHandler) HandleInput() error {
	return nil
}

func (h *UpdateUserHandler) HandleProcess() error {
	dt := data.GetData()

	// 查找用户
	user, err := data.UserNamespace.FindByUserID(dt.GetDB(), h.Req.UserID)
	if err != nil {
		log.Errorf("查找用户失败: %s", err.Error())
		h.Resp.Code = constant.ERR_USER_NOT_FOUND
		return err
	}

	// 更新用户信息
	if h.Req.Name != "" {
		user.Name = h.Req.Name
	}
	if h.Req.Nickname != "" {
		user.Nickname = h.Req.Nickname
	}
	if h.Req.Mobile != "" {
		user.Mobile = h.Req.Mobile
	}
	if h.Req.Email != "" {
		user.Email = h.Req.Email
	}
	if h.Req.LarkID != "" {
		user.LarkID = h.Req.LarkID
	}
	if h.Req.Tags != nil {
		user.Tags = data.UserTags(h.Req.Tags)
	}

	if err := data.UserNamespace.Update(dt.GetDB(), user); err != nil {
		log.Errorf("更新用户失败: %s", err.Error())
		h.Resp.Code = constant.ERR_UPDATE
		return err
	}

	log.Infof("用户更新成功: %s", h.Req.UserID)
	return nil
}

// ListUsersHandler 用户列表处理器
type ListUsersHandler struct {
	Req  ctrlmodel.ListUsersReq
	Resp ctrlmodel.ListUsersResp
}

// ListUsers 用户列表API
func ListUsers(c *gin.Context) {
	var hd ListUsersHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("ListUsers shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("ListUsers handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *ListUsersHandler) HandleInput() error {
	// 设置默认值
	if h.Req.Page <= 0 {
		h.Req.Page = 1
	}
	if h.Req.PageSize <= 0 {
		h.Req.PageSize = 10
	}
	return nil
}

func (h *ListUsersHandler) HandleProcess() error {
	dt := data.GetData()

	offset := (h.Req.Page - 1) * h.Req.PageSize
	users, total, err := data.UserNamespace.List(dt.GetDB(), offset, h.Req.PageSize)
	if err != nil {
		log.Errorf("查询用户列表失败: %s", err.Error())
		h.Resp.Code = constant.ERR_QUERY
		return err
	}

	h.Resp.Users = users
	h.Resp.Total = total
	h.Resp.Page = h.Req.Page
	return nil
}

// DeleteUserHandler 删除用户处理器
type DeleteUserHandler struct {
	Req  ctrlmodel.DeleteUserReq
	Resp ctrlmodel.DeleteUserResp
}

// DeleteUser 删除用户API
func DeleteUser(c *gin.Context) {
	var hd DeleteUserHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("DeleteUser shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("DeleteUser handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *DeleteUserHandler) HandleInput() error {
	return nil
}

func (h *DeleteUserHandler) HandleProcess() error {
	dt := data.GetData()

	if err := data.UserNamespace.Delete(dt.GetDB(), h.Req.UserID); err != nil {
		log.Errorf("删除用户失败: %s", err.Error())
		h.Resp.Code = constant.ERR_DELETE
		return err
	}

	log.Infof("用户删除成功: %s", h.Req.UserID)
	return nil
}

// FindUsersByTagsHandler 根据标签查找用户处理器
type FindUsersByTagsHandler struct {
	Req  ctrlmodel.FindUsersByTagsReq
	Resp ctrlmodel.FindUsersByTagsResp
}

// FindUsersByTags 根据标签查找用户API
func FindUsersByTags(c *gin.Context) {
	var hd FindUsersByTagsHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := c.ShouldBind(&hd.Req); err != nil {
		log.Errorf("FindUsersByTags shouldBind err %s", err.Error())
		hd.Resp.Code = constant.ERR_SHOULD_BIND
		return
	}

	if err := handler.Run(&hd); err != nil {
		log.Errorf("FindUsersByTags handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *FindUsersByTagsHandler) HandleInput() error {
	return nil
}

func (h *FindUsersByTagsHandler) HandleProcess() error {
	dt := data.GetData()

	var users []*data.User
	var err error

	if h.Req.MatchType == "all" {
		// 必须包含所有标签
		users, err = data.UserNamespace.FindByTags(dt.GetDB(), h.Req.Tags)
	} else {
		// 包含任意标签即可（默认）
		users, err = data.UserNamespace.FindByAnyTags(dt.GetDB(), h.Req.Tags)
	}

	if err != nil {
		log.Errorf("根据标签查找用户失败: %s", err.Error())
		h.Resp.Code = constant.ERR_QUERY
		return err
	}

	h.Resp.Users = users
	h.Resp.Count = len(users)
	return nil
}

// GetTagStatisticsHandler 获取标签统计处理器
type GetTagStatisticsHandler struct {
	Req  ctrlmodel.GetTagStatisticsReq
	Resp ctrlmodel.GetTagStatisticsResp
}

// GetTagStatistics 获取标签统计API
func GetTagStatistics(c *gin.Context) {
	var hd GetTagStatisticsHandler
	defer func() {
		hd.Resp.Msg = constant.GetErrMsg(hd.Resp.Code)
		c.JSON(http.StatusOK, hd.Resp)
	}()

	if err := handler.Run(&hd); err != nil {
		log.Errorf("GetTagStatistics handler.Run err %s", err.Error())
		if hd.Resp.Code == 0 {
			hd.Resp.Code = constant.ERR_INTERNAL
		}
	}
}

func (h *GetTagStatisticsHandler) HandleInput() error {
	return nil
}

func (h *GetTagStatisticsHandler) HandleProcess() error {
	dt := data.GetData()

	tagStats, err := data.UserNamespace.GetTagStatistics(dt.GetDB())
	if err != nil {
		log.Errorf("获取标签统计失败: %s", err.Error())
		h.Resp.Code = constant.ERR_QUERY
		return err
	}

	// 将map转换为数组格式
	var tagStatistics []ctrlmodel.TagStatistic
	for tag, count := range tagStats {
		tagStatistics = append(tagStatistics, ctrlmodel.TagStatistic{
			Tag:   tag,
			Count: count,
		})
	}

	h.Resp.Data = tagStatistics
	return nil
}
